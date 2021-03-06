package vm

import (
	. "luago/api"
)

/* number of list items to accumulate before a SETLIST instruction */
const LFIELDS_PER_FLUSH = 50

// newTable:R(A) := {} (size = B,C)
func newTable(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	vm.CreateTable(b, c)
	vm.Replace(a)
}

// getTable: R(A) := R(B)[RK(C)]
func getTable(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

// setTable:R(A)[Rk(B)] := RK(C) 指令的接口
func setTable(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	vm.GetRK(b) // TODO:需检查出入栈的顺序
	vm.GetRK(c)
	vm.SetTable(a)
}

// setList: R(A)[(C-1)*FPF【C表示批次数，FPF默认50】+i] := R(A+i),
//	1 <= i <= B
//  21.2.4 新增修改：当表构造器的最后一个元素是函数调用或者vararg的时候，Lua会把产生的所有值都收集起来
func setList(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	// c表示批次数，一批的大小是50，最大索引为512*50 = 25600
	// 如果数组长度>25600，c会小于等于0
	if c > 0 {
		c = c - 1
	} else {
		c = Instruction(vm.Fetch()).Ax()
	}

	bIsZero := b == 0
	if bIsZero {
		b = int(vm.ToInteger(-1)) - a - 1
		vm.Pop(1)
	}

	idx := int64(c * LFIELDS_PER_FLUSH)
	for j := 1; j <= b; j++ {
		idx++
		vm.PushValue(a + j)
		vm.SetI(a, idx)
	}

	if bIsZero {
		for j := vm.RegisterCount() + 1; j <= vm.GetTop(); j++ {
			idx++
			vm.PushValue(j)
			vm.SetI(a, idx)
		}

		// clear stack
		vm.SetTop(vm.RegisterCount())
	}
}
