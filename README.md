# Java → JVM Bytecode Compiler in Go

A compiler written in Go that translates a subset of Java source code directly into valid JVM bytecode (`.class` files). The compiled bytecode can be executed by any Java Virtual Machine.

## Project Overview

This project implements a complete three-stage compiler pipeline:

1. **Lexer** - Tokenizes Java source code
2. **Parser** - Builds an Abstract Syntax Tree (AST)
3. **Code Generator** - Emits valid JVM bytecode

It generates `.class` files that conform to the Java Class File Format specification and can be executed with the standard `java` command.

## Architecture

### Phase 1: Lexical Analysis (LEXER)

**File:** `src/lexer/lexer.go`

Converts raw Java source code into a token stream.

**Supported tokens:**

- **Keywords:** `class`, `public`, `static`, `void`, `int`, `return`
- **Operators:** `+`, `-`, `*`, `/`, `=`
- **Delimiters:** `{`, `}`, `(`, `)`, `;`, `,`, `.`, `[`, `]`
- **Literals:** Integers, Strings, Identifiers
- **Special:** Comments (skipped), Whitespace (skipped)

### Phase 2: Syntax Analysis (PARSER)

**File:** `src/parser/parser.go`

Builds an Abstract Syntax Tree from the token stream using recursive descent parsing.

**Parses:**

- Class declarations
- Method declarations (with parameters)
- Variable declarations (`int x = value;`)
- Expressions with operator precedence (Binary expressions: `+`, `-`, `*`, `/`)
- Method calls (`System.out.println(...)`)
- Return statements

### Phase 3: Code Generation (CODEGEN)

**File:** `src/compiler/codegen.go`

Generates valid JVM bytecode with:

- **Constant Pool** - Manages strings, integers, class references, method references
- **Method Code Generator** - Emits bytecode instructions (IADD, ILOAD, ISTORE, etc.)
- **Class File Writer** - Produces valid `.class` file format with proper headers

JVM bytecode instructions supported:

- `ICONST_n` - Load small integer constants
- `BIPUSH` - Push byte as integer
- `LDC` - Load from constant pool
- `ILOAD/ISTORE` - Load/store local integer variables
- `IADD/ISUB/IMUL/IDIV` - Arithmetic operations
- `INVOKEVIRTUAL` - Call instance methods
- `IRETURN/RETURN` - Return from method

**Target Version:** Java 8 (major version 52) - chosen for broad compatibility. Any newer Java version supports Java 8 bytecode.

## Supported Java Features

✅ **Working:**

- Class declarations
- Public static main methods
- Integer variables and arithmetic expressions
- Variable assignment
- Method calls (System.out.println)
- Comments and whitespace handling
- Method parameters with types
- Return statements

❌ **Not yet supported:**

- Instance variables/methods
- Inheritance
- If/else statements
- Loops (for, while)
- Object instantiation (new)
- Arrays
- Exception handling
- Generics
- Multiple classes per file

## How to Build & Run

### Prerequisites

- Go 1.25 or later
- Java JDK (to run compiled bytecode)

### Build the Compiler

```bash
go build -o compiler.exe ./src
```

### Compile Java Code

```bash
# Compile a Java file
./compiler.exe src/Example.java

# This generates: Example.class
```

### Run the Bytecode

```bash
java Example
```

## Example

**Input Java file (`Test.java`):**

```java
class Test {
    public static void main(String[] args) {
        int x = 5;
        int y = 3;
        int z = x + y;
        System.out.println(z);
    }
}
```

**Compile and run:**

```bash
./compiler.exe Test.java
java Test                    # Output: 8
```

## File Structure

```
.
├── README.md
└── src/
    ├── main.go              # Entry point
    ├── Example.java         # Test input
    ├── test_simple.java     # Minimal test
    ├── ast/
    │   └── ast.go           # AST node definitions
    ├── lexer/
    │   ├── lexer.go         # Lexical analyzer
    │   └── token.go         # Token types & keywords
    ├── parser/
    │   └── parser.go        # Syntax analyzer (recursive descent)
    └── compiler/
        └── codegen.go       # Bytecode generator
```

## Technical Details

### Constant Pool

The JVM class file format requires a constant pool—a table of constants (strings, integers, method references, etc.) that the bytecode references by index. This compiler:

- Automatically deduplicates constants
- Builds proper UTF-8, Integer, Methodref, and NameAndType entries
- Handles cross-references between pool entries

### Local Variable Management

- Variables are stored in a symbol table mapping names to local variable indices
- The code generator tracks max local variable count and adjusts ILOAD/ISTORE instructions

### Bytecode Format

The output `.class` file follows the Java Class File Format specification:

- Magic number: `0xCAFEBABE`
- Version: Java 8 (major version 52)
- Constant pool with deduplication
- Class metadata (flags, this class, super class)
- Method table with Code attributes
- Bytecode instructions and exception tables

## Future Phases

**Phase 3.5: Semantic Analysis**

- Type checking across expressions
- Variable scope validation
- Method signature verification

**Phase 4: Lowering/IR**

- Convert high-level constructs to stack machine operations
- Eliminate expression trees to single instructions

**Phase 5: Advanced Features**

- If/else conditionals (IFEQ, GOTO)
- Loops (IFEQ, GOTO jump targets)
- Instance variables and constructors
- Inheritance (extends, super calls)
- Arrays and objects

## Author

Jay