package views

type ConversationParticipant struct {
	ConversationID string
	UserID         string
}

var conversationParticipantsCols = []string{"conversation_id", "user_id"}

func (p *ConversationParticipant) values() []interface{} {
	return []interface{}{p.ConversationID, p.UserID}
}
