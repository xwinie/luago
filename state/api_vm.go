package state

// PC:获取当前的PC
func (self *luaState) PC() int {
	return self.pc
}

// AddPC:跳转到指定行指令
func (self *luaState) AddPC(n int) {
	self.pc += n
}

// Fetch: 获取当前指令，并且让指令计数器+1
func (self *luaState) Fetch() uint32 {
	i := self.proto.Code[self.pc]
	self.pc++
	return i
}

// GetConst:从函数原型中获取一个常熟并压入栈中
func (self *luaState) GetConst(idx int) {
	c := self.proto.Constants[idx]
	self.stack.push(c)
}

// GetRK:根据RK值选择将某个常量推入栈顶
//		或者调用PushValue将某个索引处的栈值推入栈顶
func (self *luaState) GetRK(rk int) {
	if rk > 0xFF {
		// 将常量值推入栈顶
		self.GetConst(rk & 0xFF)
	} else {
		self.PushValue(rk + 1)
	}
}

// 注意：GetRK的参数是OpArgK类型的，取决于9个比特中的第一个
// 虚拟机指令操作数携带的寄存器索引是从0开始的，lua栈API索引是从1开始的
// 因此寄存器索引当成栈索引使用时要+1
