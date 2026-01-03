package types

type UserAccessCheck func(openid string) *TCSError
type OnEventHandler func(event *Event)

type Custom struct {
	UserAccessCheck UserAccessCheck
	OnEventHandler  OnEventHandler
}
