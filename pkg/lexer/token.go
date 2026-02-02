package lexer

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	TokenEOF TokenType = iota
	TokenIllegal

	// Literals
	TokenIdent   // main, foo, x
	TokenInt     // 42
	TokenString  // "hello"
	TokenCharLit // 'x', '\n'

	// Keywords
	TokenInt_    // int
	TokenVoid    // void
	TokenReturn  // return
	TokenIf      // if
	TokenElse    // else
	TokenWhile   // while
	TokenDo       // do
	TokenFor      // for
	TokenBreak    // break
	TokenContinue // continue
	TokenSwitch   // switch
	TokenCase     // case
	TokenDefault  // default
	TokenGoto     // goto
	TokenTypedef  // typedef
	TokenStruct   // struct
	TokenSizeof   // sizeof
	TokenUnion    // union
	TokenEnum     // enum
	TokenStatic   // static
	TokenExtern   // extern
	TokenAuto     // auto
	TokenRegister // register
	TokenConst    // const
	TokenVolatile // volatile
	TokenRestrict  // restrict
	TokenAttribute // __attribute__
	TokenAsm       // __asm or __asm__
	TokenChar      // char
	TokenShort    // short
	TokenLong     // long
	TokenFloat    // float
	TokenDouble   // double
	TokenSigned   // signed
	TokenUnsigned // unsigned
	TokenInline   // inline, __inline, __inline__

	// Operators
	TokenPlus      // +
	TokenMinus     // -
	TokenStar      // *
	TokenSlash     // /
	TokenPercent   // %
	TokenAssign    // =
	TokenEq        // ==
	TokenNe        // !=
	TokenLt        // <
	TokenLe        // <=
	TokenGt        // >
	TokenGe        // >=
	TokenAnd       // &&
	TokenOr        // ||
	TokenNot       // !
	TokenAmpersand // &
	TokenPipe      // |
	TokenCaret     // ^
	TokenTilde     // ~
	TokenShl       // <<
	TokenShr       // >>
	TokenQuestion  // ?
	TokenColon     // :

	// Compound assignment operators
	TokenPlusAssign    // +=
	TokenMinusAssign   // -=
	TokenStarAssign    // *=
	TokenSlashAssign   // /=
	TokenPercentAssign // %=
	TokenAndAssign     // &=
	TokenOrAssign      // |=
	TokenXorAssign     // ^=
	TokenShlAssign     // <<=
	TokenShrAssign     // >>=

	// Increment/decrement
	TokenIncrement // ++
	TokenDecrement // --

	// Delimiters
	TokenLParen    // (
	TokenRParen    // )
	TokenLBrace    // {
	TokenRBrace    // }
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenSemicolon // ;
	TokenComma     // ,
	TokenDot       // .
	TokenArrow     // ->
	TokenEllipsis  // ...
)

var tokenNames = map[TokenType]string{
	TokenEOF:           "EOF",
	TokenIllegal:       "ILLEGAL",
	TokenIdent:         "IDENT",
	TokenInt:           "INT",
	TokenString:        "STRING",
	TokenCharLit:       "CHARLIT",
	TokenInt_:          "int",
	TokenVoid:          "void",
	TokenReturn:        "return",
	TokenIf:            "if",
	TokenElse:          "else",
	TokenWhile:         "while",
	TokenDo:            "do",
	TokenFor:           "for",
	TokenBreak:         "break",
	TokenContinue:      "continue",
	TokenSwitch:        "switch",
	TokenCase:          "case",
	TokenDefault:       "default",
	TokenGoto:          "goto",
	TokenTypedef:       "typedef",
	TokenStruct:        "struct",
	TokenSizeof:        "sizeof",
	TokenUnion:         "union",
	TokenEnum:          "enum",
	TokenStatic:        "static",
	TokenExtern:        "extern",
	TokenAuto:          "auto",
	TokenRegister:      "register",
	TokenConst:         "const",
	TokenVolatile:      "volatile",
	TokenRestrict:      "restrict",
	TokenAttribute:     "__attribute__",
	TokenAsm:           "__asm",
	TokenChar:          "char",
	TokenShort:         "short",
	TokenLong:          "long",
	TokenFloat:         "float",
	TokenDouble:        "double",
	TokenSigned:        "signed",
	TokenUnsigned:      "unsigned",
	TokenInline:        "inline",
	TokenPlus:          "+",
	TokenMinus:         "-",
	TokenStar:          "*",
	TokenSlash:         "/",
	TokenPercent:       "%",
	TokenAssign:        "=",
	TokenEq:            "==",
	TokenNe:            "!=",
	TokenLt:            "<",
	TokenLe:            "<=",
	TokenGt:            ">",
	TokenGe:            ">=",
	TokenAnd:           "&&",
	TokenOr:            "||",
	TokenNot:           "!",
	TokenAmpersand:     "&",
	TokenPipe:          "|",
	TokenCaret:         "^",
	TokenTilde:         "~",
	TokenShl:           "<<",
	TokenShr:           ">>",
	TokenQuestion:      "?",
	TokenColon:         ":",
	TokenPlusAssign:    "+=",
	TokenMinusAssign:   "-=",
	TokenStarAssign:    "*=",
	TokenSlashAssign:   "/=",
	TokenPercentAssign: "%=",
	TokenAndAssign:     "&=",
	TokenOrAssign:      "|=",
	TokenXorAssign:     "^=",
	TokenShlAssign:     "<<=",
	TokenShrAssign:     ">>=",
	TokenIncrement:     "++",
	TokenDecrement:     "--",
	TokenLParen:        "(",
	TokenRParen:        ")",
	TokenLBrace:        "{",
	TokenRBrace:        "}",
	TokenLBracket:      "[",
	TokenRBracket:      "]",
	TokenSemicolon:     ";",
	TokenComma:         ",",
	TokenDot:           ".",
	TokenArrow:         "->",
	TokenEllipsis:      "...",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// keywords maps keyword strings to token types
var keywords = map[string]TokenType{
	"int":      TokenInt_,
	"void":     TokenVoid,
	"return":   TokenReturn,
	"if":       TokenIf,
	"else":     TokenElse,
	"while":    TokenWhile,
	"do":       TokenDo,
	"for":      TokenFor,
	"break":    TokenBreak,
	"continue": TokenContinue,
	"switch":   TokenSwitch,
	"case":     TokenCase,
	"default":  TokenDefault,
	"goto":     TokenGoto,
	"typedef":  TokenTypedef,
	"struct":   TokenStruct,
	"sizeof":   TokenSizeof,
	"union":    TokenUnion,
	"enum":     TokenEnum,
	"static":   TokenStatic,
	"extern":   TokenExtern,
	"auto":     TokenAuto,
	"register": TokenRegister,
	"const":    TokenConst,
	"volatile": TokenVolatile,
	"restrict":       TokenRestrict,
	"__attribute__":  TokenAttribute,
	"__asm":          TokenAsm,
	"__asm__":        TokenAsm,
	"char":           TokenChar,
	"short":    TokenShort,
	"long":     TokenLong,
	"float":    TokenFloat,
	"double":   TokenDouble,
	"signed":     TokenSigned,
	"unsigned":   TokenUnsigned,
	"inline":     TokenInline,
	"__inline":   TokenInline,
	"__inline__": TokenInline,
}

// LookupIdent returns the token type for an identifier (keyword or IDENT)
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdent
}
