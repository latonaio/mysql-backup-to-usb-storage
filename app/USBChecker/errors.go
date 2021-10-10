package USBChecker

type MountError struct {
	message string
}

func (e *MountError) Error() string {
	return e.message
}

type NotConnect struct {
	message string
}

func (e *NotConnect) Error() string {
	return e.message
}
