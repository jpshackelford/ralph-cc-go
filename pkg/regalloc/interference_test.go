package regalloc

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/rtl"
)

func TestInterferenceGraphBasic(t *testing.T) {
	g := NewInterferenceGraph()

	// Add some nodes
	g.AddNode(1)
	g.AddNode(2)
	g.AddNode(3)

	if !g.Nodes.Contains(1) || !g.Nodes.Contains(2) || !g.Nodes.Contains(3) {
		t.Error("nodes should be added")
	}

	// Add edges
	g.AddEdge(1, 2)
	g.AddEdge(1, 3)

	if !g.HasEdge(1, 2) || !g.HasEdge(2, 1) {
		t.Error("edge 1-2 should exist (bidirectional)")
	}
	if !g.HasEdge(1, 3) || !g.HasEdge(3, 1) {
		t.Error("edge 1-3 should exist (bidirectional)")
	}
	if g.HasEdge(2, 3) {
		t.Error("edge 2-3 should not exist")
	}
}

func TestInterferenceGraphDegree(t *testing.T) {
	g := NewInterferenceGraph()

	g.AddEdge(1, 2)
	g.AddEdge(1, 3)
	g.AddEdge(1, 4)
	g.AddEdge(2, 3)

	if d := g.Degree(1); d != 3 {
		t.Errorf("degree of 1 = %d, want 3", d)
	}
	if d := g.Degree(2); d != 2 {
		t.Errorf("degree of 2 = %d, want 2", d)
	}
	if d := g.Degree(3); d != 2 {
		t.Errorf("degree of 3 = %d, want 2", d)
	}
	if d := g.Degree(4); d != 1 {
		t.Errorf("degree of 4 = %d, want 1", d)
	}
}

func TestInterferenceGraphNeighbors(t *testing.T) {
	g := NewInterferenceGraph()

	g.AddEdge(1, 2)
	g.AddEdge(1, 3)

	neighbors := g.Neighbors(1)
	if !neighbors.Contains(2) || !neighbors.Contains(3) {
		t.Error("neighbors of 1 should be {2, 3}")
	}
	if len(neighbors) != 2 {
		t.Errorf("neighbors of 1 has %d elements, want 2", len(neighbors))
	}
}

func TestInterferenceGraphRemoveNode(t *testing.T) {
	g := NewInterferenceGraph()

	g.AddEdge(1, 2)
	g.AddEdge(1, 3)
	g.AddEdge(2, 3)
	g.AddPreference(1, 2)

	g.RemoveNode(1)

	if g.Nodes.Contains(1) {
		t.Error("node 1 should be removed")
	}
	if g.HasEdge(1, 2) || g.HasEdge(2, 1) {
		t.Error("edges to 1 should be removed")
	}
	// Node 2-3 edge should remain
	if !g.HasEdge(2, 3) {
		t.Error("edge 2-3 should remain")
	}
}

func TestInterferenceGraphPreferences(t *testing.T) {
	g := NewInterferenceGraph()

	g.AddPreference(1, 2)
	g.AddPreference(1, 3)

	if !g.Preferences[1].Contains(2) || !g.Preferences[1].Contains(3) {
		t.Error("preferences of 1 should include 2 and 3")
	}
	if !g.Preferences[2].Contains(1) {
		t.Error("preferences should be bidirectional")
	}
}

func TestInterferenceGraphMoveRelated(t *testing.T) {
	g := NewInterferenceGraph()

	g.AddNode(1)
	g.AddNode(2)
	g.AddNode(3)

	g.AddPreference(1, 2)

	if !g.MoveRelated(1) {
		t.Error("1 should be move-related")
	}
	if !g.MoveRelated(2) {
		t.Error("2 should be move-related")
	}
	if g.MoveRelated(3) {
		t.Error("3 should not be move-related")
	}
}

func TestBuildInterferenceGraphSimple(t *testing.T) {
	// Simple function:
	// 1: x1 = int 1
	// 2: x2 = int 2
	// 3: x3 = add(x1, x2)
	// 4: return x3
	fn := &rtl.Function{
		Name: "simple",
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Iop{Op: rtl.Ointconst{Value: 1}, Args: nil, Dest: 1, Succ: 2},
			2: rtl.Iop{Op: rtl.Ointconst{Value: 2}, Args: nil, Dest: 2, Succ: 3},
			3: rtl.Iop{Op: rtl.Oadd{}, Args: []rtl.Reg{1, 2}, Dest: 3, Succ: 4},
			4: rtl.Ireturn{Arg: ptr(rtl.Reg(3))},
		},
		Entrypoint: 1,
	}

	liveness := AnalyzeLiveness(fn)
	g := BuildInterferenceGraph(fn, liveness)

	// x1 and x2 should interfere (both live when x2 is defined)
	if !g.HasEdge(1, 2) {
		t.Error("x1 and x2 should interfere")
	}

	// x3 should not interfere with x1 or x2 (they're dead when x3 is defined)
	if g.HasEdge(3, 1) {
		t.Error("x3 and x1 should not interfere")
	}
	if g.HasEdge(3, 2) {
		t.Error("x3 and x2 should not interfere")
	}
}

func TestBuildInterferenceGraphWithMove(t *testing.T) {
	// Function with move:
	// 1: x1 = int 42
	// 2: x2 = move(x1)
	// 3: return x2
	fn := &rtl.Function{
		Name: "move",
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Iop{Op: rtl.Ointconst{Value: 42}, Args: nil, Dest: 1, Succ: 2},
			2: rtl.Iop{Op: rtl.Omove{}, Args: []rtl.Reg{1}, Dest: 2, Succ: 3},
			3: rtl.Ireturn{Arg: ptr(rtl.Reg(2))},
		},
		Entrypoint: 1,
	}

	liveness := AnalyzeLiveness(fn)
	g := BuildInterferenceGraph(fn, liveness)

	// x1 and x2 should NOT interfere (move coalescing)
	// This is the key optimization: moves don't create interference
	if g.HasEdge(1, 2) {
		t.Error("x1 and x2 should NOT interfere (move instruction)")
	}

	// But they should have a preference
	if !g.Preferences[1].Contains(2) || !g.Preferences[2].Contains(1) {
		t.Error("x1 and x2 should have preference edge")
	}
}

func TestBuildInterferenceGraphHighPressure(t *testing.T) {
	// Function with high register pressure:
	// All registers live at the same time
	// 1: x1 = int 1
	// 2: x2 = int 2
	// 3: x3 = int 3
	// 4: x4 = add(x1, x2)
	// 5: x5 = add(x3, x4)
	// 6: return x5
	fn := &rtl.Function{
		Name: "pressure",
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Iop{Op: rtl.Ointconst{Value: 1}, Args: nil, Dest: 1, Succ: 2},
			2: rtl.Iop{Op: rtl.Ointconst{Value: 2}, Args: nil, Dest: 2, Succ: 3},
			3: rtl.Iop{Op: rtl.Ointconst{Value: 3}, Args: nil, Dest: 3, Succ: 4},
			4: rtl.Iop{Op: rtl.Oadd{}, Args: []rtl.Reg{1, 2}, Dest: 4, Succ: 5},
			5: rtl.Iop{Op: rtl.Oadd{}, Args: []rtl.Reg{3, 4}, Dest: 5, Succ: 6},
			6: rtl.Ireturn{Arg: ptr(rtl.Reg(5))},
		},
		Entrypoint: 1,
	}

	liveness := AnalyzeLiveness(fn)
	g := BuildInterferenceGraph(fn, liveness)

	// At node 4, x1, x2, x3 are all live
	// So they should all interfere with each other
	if !g.HasEdge(1, 2) {
		t.Error("x1-x2 should interfere")
	}
	if !g.HasEdge(1, 3) {
		t.Error("x1-x3 should interfere")
	}
	if !g.HasEdge(2, 3) {
		t.Error("x2-x3 should interfere")
	}
}

func TestNoSelfEdges(t *testing.T) {
	g := NewInterferenceGraph()

	g.AddEdge(1, 1)
	if g.HasEdge(1, 1) {
		t.Error("self-edges should not be added")
	}

	g.AddPreference(2, 2)
	if g.Preferences[2].Contains(2) {
		t.Error("self-preferences should not be added")
	}
}
