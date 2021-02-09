package state

import (
	. "luago/api"
	"luago/number"
	"math"
)

type operator struct {
	metamethod  string                         // 元方法
	integerFunc func(int64, int64) int64       // 整型运算
	floatFunc   func(float64, float64) float64 // 浮点型运算
}

var (
	iadd  = func(a, b int64) int64 { return a + b }
	fadd  = func(a, b float64) float64 { return a + b }
	isub  = func(a, b int64) int64 { return a - b }
	fsub  = func(a, b float64) float64 { return a - b }
	imul  = func(a, b int64) int64 { return a * b }
	fmul  = func(a, b float64) float64 { return a * b }
	imod  = number.IMod
	fmod  = number.FMod
	pow   = math.Pow
	div   = func(a, b float64) float64 { return a / b }
	iidiv = number.IFloorDiv
	fidiv = number.FFloorDiv
	band  = func(a, b int64) int64 { return a & b }
	bor   = func(a, b int64) int64 { return a | b }
	bxor  = func(a, b int64) int64 { return a ^ b }
	shl   = number.ShiftLeft
	shr   = number.ShiftRight
	iunm  = func(a, _ int64) int64 { return -a }
	funm  = func(a, _ float64) float64 { return -a }
	bnot  = func(a, _ int64) int64 { return ^a }
)

var operators = []operator{
	operator{"__add", iadd, fadd},
	operator{"__sub", isub, fsub},
	operator{"__mul", imul, fmul},
	operator{"__mod", imod, fmod},
	operator{"__pow", nil, pow},
	operator{"__div", nil, div},
	operator{"__idiv", iidiv, fidiv},
	operator{"__band", band, nil},
	operator{"__bor", bor, nil},
	operator{"__bxor", bxor, nil},
	operator{"__shl", shl, nil},
	operator{"__shr", shr, nil},
	operator{"__unm", iunm, funm},
	operator{"__bnot", bnot, nil},
}

// Arith:执行计算
func (self *luaState) Arith(op ArithOp) {
	var a, b luaValue
	b = self.stack.pop()
	if op != LUA_OPUNM && op != LUA_OPBNOT {
		a = self.stack.pop()
	} else {
		// 针对只有一个操作数的运算符
		a = b
	}
	// 这里作者好像有个笔误 应该是先取出a操作数，再取出b操作数
	// 更正 并非作者笔误，起初在写abc模式提取操作数时，误认为排列时cba，
	// 	实际上的排列是 B(9位) C(9位) A(8位) OPCODE(6位)
	operator := operators[op]
	if result := _arith(a, b, operator); result != nil {
		self.stack.push(result)
		return
	}
	mm := operator.metamethod
	if result, ok := callMetamethod(a, b, mm, self); ok {
		self.stack.push(result)
		return
	}
	panic("arithmetic error")
}

// _arith: 区分类型的辅助函数
func _arith(a, b luaValue, op operator) luaValue {
	if op.floatFunc == nil {
		// 没有浮点型运算函数一般为位运算，先将其转换为整数
		if x, ok := convertToInteger(a); ok {
			if y, ok := convertToInteger(b); ok {
				return op.integerFunc(x, y)
			}
		}
	} else {
		// 常规算术符运算
		if op.integerFunc != nil {
			// add,sub,mul,mod,idiv,unm
			if x, ok := a.(int64); ok {
				if y, ok := b.(int64); ok {
					return op.integerFunc(x, y)
				}
			}
		}
		if x, ok := convertToFloat(a); ok {
			if y, ok := convertToFloat(b); ok {
				return op.floatFunc(x, y)
			}
		}
	}
	return nil
}
