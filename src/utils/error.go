package utils

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicIfErrWithOutput(output string, err error) {
	if err != nil {
		panic(output)
	}
}
