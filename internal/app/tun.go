package app

import "github.com/metacubex/mihomo/constant/features"

// TunStack is a network stack the TUN can use.
type TunStack struct {
	// Value is what the configuration says.
	Value string
	// Name is what the interface calls it.
	Name string
}

// tunStacks are the stacks the TUN can use, in the order the interface
// shows them.
var tunStacks = []TunStack{
	{Value: "gvisor", Name: "GVisor"},
	{Value: "system", Name: "系统"},
	{Value: "mixed", Name: "混合"},
}

// TunStackAvailable reports whether the core of this build can use the
// stack. The gVisor stack and the mixed one are only in the build when it
// was compiled with the with_gvisor tag; without it, the core says so and
// the TUN stays down.
func TunStackAvailable(stack string) bool {
	switch stack {
	case "gvisor", "mixed":
		return features.WithGVisor
	}
	return true
}

// TunStacks returns the stacks this build can use.
func TunStacks() []TunStack {
	out := make([]TunStack, 0, len(tunStacks))
	for _, s := range tunStacks {
		if TunStackAvailable(s.Value) {
			out = append(out, s)
		}
	}
	return out
}

// normalizeTunStack keeps a stack the core of this build can use: a build
// without the gVisor stack falls back to the system one.
func normalizeTunStack(stack string) string {
	for _, s := range tunStacks {
		if s.Value == stack {
			if TunStackAvailable(stack) {
				return stack
			}
			return "system"
		}
	}
	return "system"
}
