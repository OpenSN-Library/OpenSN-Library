package utils

func DoWithRetry(thing func() error, maxTime int) error {
	var err error
	for i := 0; i < maxTime; i++ {
		err = thing()
		if err == nil {
			break
		}
	}
	return err
}
