package vm

func (m *mockQMPClient) QueryNetdev() ([]QMPNetDeviceStats, error) {
	if m.qError != nil {
		return nil, m.qError
	}
	return nil, nil
}

func (c *countingBlockClient) QueryNetdev() ([]QMPNetDeviceStats, error) { return nil, nil }
