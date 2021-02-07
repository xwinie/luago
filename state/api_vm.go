package state

// PC:获取当前的PC
func (self *luaState) PC() int {
	return self.stack.pc
}

// AddPC:跳转到指定行指令
func (self *luaState) AddPC(n int) {
	self.stack.pc += n
}

// Fetch: 获取当前指令，并且让指令计数器+1
func (self *luaState) Fetch() uint32 {
	i := self.stack.closure.proto.Code[self.stack.pc]
	self.stack.pc++
	return i
}

// GetConst:从函数原型中获取一个常熟并压入栈中
func (self *luaState) GetConst(idx int) {
	c := self.stack.closure.proto.Constants[idx]
	self.stack.push(c)
}

// GetRK:根据RK值选择将某个常量推入栈顶
//		或者调用PushValue将某个索引处的栈值推入栈顶
func (self *luaState) GetRK(rk int) {
	if rk > 0xFF {
		// 将常量值推入栈顶
		self.GetConst(rk & 0xFF)
	} else {
		// rk相关的在这里已经+1了无需处理
		self.PushValue(rk + 1)
	}
}

// 注意：GetRK的参数是OpArgK类型的，取决于9个比特中的第一个
// 虚拟机指令操作数携带的寄存器索引是从0开始的，lua栈API索引是从1开始的
// 因此寄存器索引当成栈索引使用时要+1

// RegisterCount:返回当前Lua函数所操作的寄存器数量
func (self *luaState) RegisterCount() int {
	return int(self.stack.closure.proto.MaxStackSize)
}

//LoadVararg(n int):传递给当前Lua函数的变长参数推入栈顶
func (self *luaState) LoadVararg(n int) {
	if n < 0 {
		n = len(self.stack.varargs)
	}
	self.stack.check(n)
	self.stack.pushN(self.stack.varargs, n)
}

//LoadProto(idx int):把当前Lua函数的子函数原型 实例化为闭包推入栈顶
func (self *luaState) LoadProto(idx int) {
	stack := self.stack
	subProto := stack.closure.proto.Protos[idx] // 栈内的外部原型
	closure := newLuaClosure(subProto)
	self.stack.push(closure)
	//加载UpValue
	for i, uvInfo := range subProto.Upvalues {
		uvIdx := int(uvInfo.Idx)
		if uvInfo.Instack == 1 {
			// 如果某个Upvalue捕获的是当前函数的局部变量，只要访问当前函数的局部变量即可
			if stack.openuvs == nil {
				stack.openuvs = map[int]*upvalue{}
			}
			if openuv, found := stack.openuvs[uvIdx]; found {
				closure.upvals[i] = openuv
			} else {
				closure.upvals[i] = &upvalue{&stack.slots[uvIdx]}
				// 加载子函数原型时会将栈内的upvalue值放入当前的openuv
				stack.openuvs[uvIdx] = closure.upvals[i]
			}
		} else {
			// 更外围函数的局部变量
			closure.upvals[i] = stack.closure.upvals[uvIdx]
		}
	}

}

// CloseUpvalues:关闭openuv的通道
//	如果某个块内部定义的局部变量已经被嵌套函数捕获，那么当这些局部变量退出作用域时（块结束）
//	编译器会产生一条jmp指令，指示虚拟机闭合相应的Upvalue
func (self *luaState) CloseUpvalues(a int) {
	for i, openuv := range self.stack.openuvs {
		if i >= a-1 {
			val := *openuv.val
			openuv.val = &val
			delete(self.stack.openuvs, i)
		}
	}
}
