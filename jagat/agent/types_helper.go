package agent

// helper

func NewTextMessage(role Role, text string) *Message {
	m := &Message{
		Role: role,
		Parts: []*Part{
			{Text: text},
		},
	}
	return m
}

func NewBlobMessage(role Role, b []byte, mime string) *Message {
	m := Message{
		Role: role,
		Parts: []*Part{
			{
				Blob: &Blob{
					Bytes: b,
					Mime:  mime,
				},
			},
		},
	}
	return &m
}
