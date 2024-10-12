# wat-go: WebAssembly文本格式工具

目标不是解析完整的 wat 语法, 而是能满足wat-go输出的 wat 格式.

- 安装`wat-go`命令行: `go install github.com/chai2010/wat-go@master`
- `wat-go strip`子命令: 减小wat文件体积, 只保留导出对象和依赖的代码
- `wat-go 2c`子命令: 将 wat 代码转为 C 代码

## Wat 格式的子集

- 函数指令不支持折叠
- 只支持行注释, 不支持多行块注释
- 每个指令一行, 单指令之间不会出现行注释
- 对象前出现的是关联注释, 其他注释全部丢弃
- 转义字符串扩展: '\n', '\r', '\t', '\\', '\"'

fib.wat 代码如下:

```wat
(module
  (func $foo)
  (func $add (export "add") (param $a i64) (param $b i64) (result i64)
    local.get $a
    local.get $b
    i64.add
  )
  (func $fib (export "fib") (param $n i64) (result i64)
    local.get $n
    i64.const 2
    i64.le_u
    if
      i64.const 1
      return
    end
    local.get $n
    i64.const 1
    i64.sub
    call $fib
    local.get $n
    i64.const 2
    i64.sub
    call $fib
    i64.add
  )
)
```

## 例子：Fib 瘦身

执行以下命令：

```
$ wat-go strip fib.wat
```

将删除代码中的 `$foo` 函数。

## 例子：Fib 转为 C 代码

输入以下命令转为C语言代码:

```
$ wat-go 2c fib.wat
```

输出的`_a.out.c`内容如下:

```c
// Auto Generated by https://github.com/chai2010/wat-go. DONOT EDIT!!!

#include <stdint.h>
#include <string.h>
#include <math.h>

typedef uint8_t   u8_t;
typedef int8_t    i8_t;
typedef uint16_t  u16_t;
typedef int16_t   i16_t;
typedef uint32_t  u32_t;
typedef int32_t   i32_t;
typedef uint64_t  u64_t;
typedef int64_t   i64_t;
typedef float     f32_t;
typedef double    f64_t;
typedef uintptr_t ref_t;

typedef union val_t {
  i64_t i64;
  f64_t f64;
  i32_t i32;
  f32_t f32;
  ref_t ref;
} val_t;

// func $foo
static void fn_foo();
// func $add (param $a i64) (param $b i64) (result i64)
static i64_t fn_add(i64_t a, i64_t b);
// func $fib (param $n i64) (result i64)
static i64_t fn_fib(i64_t n);

// func foo
static void fn_foo() {
  u32_t $R_u32;
  u16_t $R_u16;
  u8_t  $R_u8;
  val_t $R0;

}

// func add (param $a i64) (param $b i64) (result i64)
static i64_t fn_add(i64_t a, i64_t b) {
  u32_t $R_u32;
  u16_t $R_u16;
  u8_t  $R_u8;
  val_t $R0, $R1;

  $R0.i64 = a;
  $R1.i64 = b;
  $R0.i64 = $R0.i64 + $R1.i64;
  return $R0.i64;
}

// func fib (param $n i64) (result i64)
static i64_t fn_fib(i64_t n) {
  u32_t $R_u32;
  u16_t $R_u16;
  u8_t  $R_u8;
  val_t $R0, $R1, $R2;

  $R0.i64 = n;
  $R1.i64 = 2;
  $R0.i32 = ((u64_t)($R0.i64)<=(u64_t)($R1.i64))? 1: 0;
  if($R0.i32) {
    $R0.i64 = 1;
    return $R0.i64;
  }
  $R0.i64 = n;
  $R1.i64 = 1;
  $R0.i64 = $R0.i64 - $R1.i64;
  $R0.i64 = fn_fib($R0.i64);
  $R1.i64 = n;
  $R2.i64 = 2;
  $R1.i64 = $R1.i64 - $R2.i64;
  $R1.i64 = fn_fib($R1.i64);
  $R0.i64 = $R0.i64 + $R1.i64;
  return $R0.i64;
}
```

## 输出C代码的性能

进入`testdata/bench/wat2c`目录执行以下命令：

```
$ make
clang -O0 -o fib_c_native_O0.exe _fib_c_native.c
wat-go 2c -o fib_wat2c_native.c fib_wat.txt && clang -O0 -o fib_wat2c_native_O0.exe fib_wat2c_main.c
clang -O1 -o fib_c_native_O1.exe _fib_c_native.c
wat-go 2c -o fib_wat2c_native.c fib_wat.txt && clang -O1 -o fib_wat2c_native_O1.exe fib_wat2c_main.c
clang -O3 -o fib_c_native_O3.exe _fib_c_native.c
wat-go 2c -o fib_wat2c_native.c fib_wat.txt && clang -O3 -o fib_wat2c_native_O3.exe fib_wat2c_main.c
time ./fib_c_native_O0.exe
fib(46) = 1836311903
       11.70 real         9.87 user         0.09 sys
time ./fib_wat2c_native_O0.exe
fib(46) = 1836311903
       31.30 real        27.55 user         0.21 sys
time ./fib_c_native_O1.exe
fib(46) = 1836311903
        4.88 real         4.84 user         0.01 sys
time ./fib_wat2c_native_O1.exe
fib(46) = 1836311903
        5.54 real         4.92 user         0.04 sys
time ./fib_c_native_O3.exe
fib(46) = 1836311903
        5.68 real         5.08 user         0.04 sys
time ./fib_wat2c_native_O3.exe
fib(46) = 1836311903
        5.06 real         4.81 user         0.01 sys
```

wat转译到C代码在`-O1`和`-O3`优化的执行性能和本地C版本持平.
