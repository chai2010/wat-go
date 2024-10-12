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
