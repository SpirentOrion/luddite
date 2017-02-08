package logrus

type Suppressor interface {
	ShouldSuppress(entry *Entry) bool
}
