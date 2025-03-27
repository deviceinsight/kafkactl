package producer

type DefaultMessageSerializer struct {
	topic string
}

func (serializer DefaultMessageSerializer) CanSerializeValue(_ string) (bool, error) {
	return true, nil
}

func (serializer DefaultMessageSerializer) CanSerializeKey(_ string) (bool, error) {
	return true, nil
}

func (serializer DefaultMessageSerializer) SerializeValue(value []byte, _ Flags) ([]byte, error) {
	return value, nil
}

func (serializer DefaultMessageSerializer) SerializeKey(key []byte, _ Flags) ([]byte, error) {
	return key, nil
}
