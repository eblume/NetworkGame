package util

// TODO - Way more stuff should be in here. Stuff to reduce boilerplate, for
// instance. It's kind of rediculous. If nothing can be done to add to this,
// then the entire idea of this being a module should probably be scrapped.

type Call struct {
	Args []interface{}
	Done chan interface{}
}
