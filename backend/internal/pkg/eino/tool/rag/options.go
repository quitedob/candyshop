package rag

// Option customizes retriever behavior.
type Option func(*Options)

// Options controls corpus parsing and retrieval behavior.
type Options struct {
	ChunkSize        int
	ChunkOverlapWord int
	DefaultTopK      int
	MaxTopK          int
	MinScore         float64
	StrictCorpus     bool
}

func defaultOptions() Options {
	return Options{
		ChunkSize:        900,
		ChunkOverlapWord: 20,
		DefaultTopK:      3,
		MaxTopK:          8,
		MinScore:         0.08,
		StrictCorpus:     false,
	}
}

func applyOptions(opts ...Option) Options {
	cfg := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = defaultOptions().ChunkSize
	}
	if cfg.ChunkOverlapWord < 0 {
		cfg.ChunkOverlapWord = defaultOptions().ChunkOverlapWord
	}
	if cfg.DefaultTopK <= 0 {
		cfg.DefaultTopK = defaultOptions().DefaultTopK
	}
	if cfg.MaxTopK <= 0 {
		cfg.MaxTopK = defaultOptions().MaxTopK
	}
	if cfg.MaxTopK < cfg.DefaultTopK {
		cfg.MaxTopK = cfg.DefaultTopK
	}
	if cfg.MinScore < 0 {
		cfg.MinScore = 0
	}
	return cfg
}

// WithChunkSize sets the maximum chunk size in characters.
func WithChunkSize(size int) Option {
	return func(o *Options) {
		o.ChunkSize = size
	}
}

// WithChunkOverlapWord sets overlap words between long chunks.
func WithChunkOverlapWord(overlap int) Option {
	return func(o *Options) {
		o.ChunkOverlapWord = overlap
	}
}

// WithDefaultTopK sets the default retrieval TopK value.
func WithDefaultTopK(topK int) Option {
	return func(o *Options) {
		o.DefaultTopK = topK
	}
}

// WithMaxTopK sets the maximum allowed TopK value.
func WithMaxTopK(topK int) Option {
	return func(o *Options) {
		o.MaxTopK = topK
	}
}

// WithMinScore sets the minimum score threshold for retrieval matches.
func WithMinScore(score float64) Option {
	return func(o *Options) {
		o.MinScore = score
	}
}

// WithStrictCorpus enables strict corpus parsing mode.
func WithStrictCorpus(strict bool) Option {
	return func(o *Options) {
		o.StrictCorpus = strict
	}
}
