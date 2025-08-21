package generator

// Options is the options for the retriever.
type Options struct {
}

type Option struct {
	apply func(opts *Options)

	implSpecificOptFn any
}
