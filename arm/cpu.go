package arm

import (
	"ndsemu/emu"

	log "gopkg.in/Sirupsen/logrus.v0"
)

type Arch int
type Line int

const (
	// NOTE: the order is important. Code can do things like "if arch <) ARMv4"
	// to mean "ARMv4 and earlier"
	ARMv4 Arch = 4
	ARMv5 Arch = 5
)

const (
	LineFiq Line = 1 << iota
	LineIrq
	LineHalt
)

type Cpu struct {
	Regs  [16]reg
	Cpsr  regCpsr
	Clock int64

	UsrBank  [2]reg
	FiqBank  [2]reg
	SvcBank  [2]reg
	AbtBank  [2]reg
	IrqBank  [2]reg
	UndBank  [2]reg
	SpsrBank [5]reg

	UsrBank2 [5]reg
	FiqBank2 [5]reg

	arch  Arch
	bus   emu.Bus
	pc    reg
	cp15  *Cp15
	cops  [16]Coprocessor
	lines Line

	// Store the previous PC, used for debugging (eg: jumping into nowhere)
	prevpc reg

	// Number of cycles consumed when accessing the external bus
	memCycles int64

	// manual tracing support
	DebugTrace int
}

func NewCpu(arch Arch, bus emu.Bus) *Cpu {
	cpu := &Cpu{bus: bus, arch: arch}
	cpu.Cpsr.r = 0x13 // mode supervisor
	cpu.memCycles = int64(bus.WaitStates() + 1)
	return cpu
}

func (cpu *Cpu) SetPC(addr uint32) {
	cpu.Regs[15] = reg(addr)
}

func (cpu *Cpu) RegSpsrForMode(mode CpuMode) *reg {
	switch mode {
	case CpuModeUser, CpuModeSystem:
		log.Fatalf("non-privileged mode in RegSpsr(): %v", mode)
		panic("unreachable")
	case CpuModeFiq:
		return &cpu.SpsrBank[0]
	case CpuModeSupervisor:
		return &cpu.SpsrBank[1]
	case CpuModeAbort:
		return &cpu.SpsrBank[2]
	case CpuModeIrq:
		return &cpu.SpsrBank[3]
	case CpuModeUndefined:
		return &cpu.SpsrBank[4]
	default:
		log.Fatalf("unsupported mode in RegSpsr(): %v", mode)
		panic("unreachable")
	}
}

func (cpu *Cpu) RegF14ForMode(mode CpuMode) *reg {
	switch mode {
	case CpuModeUser, CpuModeSystem:
		return &cpu.UsrBank[1]
	case CpuModeFiq:
		return &cpu.FiqBank[1]
	case CpuModeSupervisor:
		return &cpu.SvcBank[1]
	case CpuModeAbort:
		return &cpu.AbtBank[1]
	case CpuModeIrq:
		return &cpu.IrqBank[1]
	case CpuModeUndefined:
		return &cpu.UndBank[1]
	default:
		log.Fatalf("unsupported mode in RegSpsr(): %v", mode)
		panic("unreachable")
	}
}

func (cpu *Cpu) RegSpsr() *reg {
	return cpu.RegSpsrForMode(cpu.Cpsr.GetMode())
}

func (cpu *Cpu) MapCoprocessor(copnum int, cop Coprocessor) {
	cpu.cops[copnum] = cop
}

func (cpu *Cpu) EnableCp15() *Cp15 {
	cpu.cp15 = newCp15(cpu)
	cpu.cops[15] = cpu.cp15
	return cpu.cp15
}

type Exception int

const (
	ExceptionReset           Exception = 0
	ExceptionUndefined       Exception = 1
	ExceptionSwi             Exception = 2
	ExceptionPrefetchAbort   Exception = 3
	ExceptionDataAbort       Exception = 4
	ExceptionAddressOverflow Exception = 5
	ExceptionIrq             Exception = 6
	ExceptionFiq             Exception = 7
)

// CPU mode to enter when the exception is raised
var excMode = [8]CpuMode{
	CpuModeSupervisor,
	CpuModeUndefined,
	CpuModeSupervisor,
	CpuModeAbort,
	CpuModeAbort,
	CpuModeSupervisor,
	CpuModeIrq,
	CpuModeFiq,
}

var excPcOffsetArm = [8]uint32{
	0, 0, 0, 0, 4, 0, 0, 0,
}
var excPcOffsetThumb = [8]uint32{
	0, 0, 0, 2, 4, 0, 2, 2,
}

func (cpu *Cpu) Exception(exc Exception) {
	newmode := excMode[exc]

	// Check if FIQ/IRQ are disabled
	if exc == ExceptionFiq && cpu.Cpsr.F() {
		return
	}
	if exc == ExceptionIrq && cpu.Cpsr.I() {
		return
	}

	pc := cpu.pc
	if cpu.Cpsr.T() {
		pc += reg(excPcOffsetThumb[exc])
	} else {
		pc += reg(excPcOffsetArm[exc])
	}
	*cpu.RegSpsrForMode(newmode) = cpu.Cpsr.r
	*cpu.RegF14ForMode(newmode) = pc
	cpu.Cpsr.SetT(false)
	cpu.Cpsr.SetWithMask(uint32(newmode), 0x1F, cpu)
	cpu.Cpsr.SetI(true)
	if exc == ExceptionReset || exc == ExceptionFiq {
		cpu.Cpsr.SetF(true)
	}

	if cpu.cp15 != nil {
		cpu.Regs[15] = reg(cpu.cp15.ExceptionVector())
	} else {
		cpu.Regs[15] = 0x00000000
	}

	cpu.Regs[15] += reg(exc * 4)
	if exc == ExceptionSwi {
		log.Infof("SWI 0x%x: LR=%v, arch=%v", cpu.opRead16(uint32(pc-2))&0xFF, pc, cpu.arch)
	} else {
		log.Warnf("Exception: exc=%v, LR=%v, arch=%v", exc, pc, cpu.arch)
	}
	cpu.pc = cpu.Regs[15]
	cpu.lines &^= LineHalt
}

// Set the status of the external (virtual) lines. This is modeled
// to resemble the physical lines of the CPU core, but without the
// need of full fidelity to high/low signals or clocking.
//
// For virtual lines, "true" means "activate the function", while
// "false" means "disable the function" (irrespecitve of the physical
// high/low signal required by the core).
func (cpu *Cpu) SetLine(line Line, val bool) {
	if val {
		cpu.lines |= line
	} else {
		cpu.lines &^= line
	}
}

func (cpu *Cpu) Reset() {
	cpu.Exception(ExceptionReset)
}
