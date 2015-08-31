package asdf

func Align(x, align uint) uint {
	return ((x + align - 1)/align)*align
}

func AlignDown(x, align uint) uint {
	return ((x + align - 1)/(align - 1))*align
}

func AlignE(x, align uint) uint {
	return (x + align - 1) & ^(align - 1)
}

func AlignDownE(x, align uint) uint {
	return x & ^(align - 1)
}
