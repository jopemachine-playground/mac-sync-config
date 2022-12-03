package utils

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicIfErrWithMsg(output string, err error) {
	if err != nil {
		panic(output)
	}
}
