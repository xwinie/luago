package vm

import (
	. "luago/api"
)

// forPrep:R(A) -= R(A+2); pc += sBx
// 循环开始前预先给数值减去步场，然后跳转到FORLOOP指令开始循环
func forPrep(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	if vm.Type(a) == LUA_TSTRING {
		vm.PushNumber(vm.ToNumber(a))
		vm.Replace(a)
	}
	if vm.Type(a+1) == LUA_TSTRING {
		vm.PushNumber(vm.ToNumber(a + 1))
		vm.Replace(a + 1)
	}
	if vm.Type(a+2) == LUA_TSTRING {
		vm.PushNumber(vm.ToNumber(a + 2))
		vm.Replace(a + 2)
	}

	// R(A) -= R(A+2)
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(LUA_OPSUB)
	vm.Replace(a)
	// pc += sBx
	vm.AddPC(sBx)
}

// forLoop: R(A) += R(A+2) if R(A) <= R(A+1) then pc+=sBx; R(A+3) = R(A)
//	和ForPrep不一样，先给数值加上步场，然后判断是否在范围内，再执行循环体内的代码
func forLoop(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	// R(A) += R(A+2)
	vm.PushValue(a + 2)
	vm.PushValue(a)
	vm.Arith(LUA_OPADD)
	vm.Replace(a)

	// R(A) <?= R(A+1)
	isPositiveStep := vm.ToNumber(a+2) >= 0
	if isPositiveStep && vm.Compare(a, a+1, LUA_OPLE) ||
		!isPositiveStep && vm.Compare(a+1, a, LUA_OPLE) {
		vm.AddPC(sBx)
		vm.Copy(a, a+3)
	}
}
