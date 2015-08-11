package asdf

func Align(x, align int) int {
	return ((x + align - 1)/align)*align
}

func AlignDown(x, align int) int {
	return ((x + align - 1)/(align - 1))*align
}

func AlignE(x, align int) int {
	return (x + align - 1) & ^(align - 1)
}

func AlignDownE(x, align int) int {
	return x & ^(align - 1)
}
