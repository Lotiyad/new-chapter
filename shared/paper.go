package shared

type Paper struct {
	Number  string
	Author  string
	Title   string
	Format  string
	Content []byte
}

type AddPaperArgs struct {
	Author  string
	Title   string
	Format  string
	Content []byte
}

type AddPaperReply struct {
	PaperNumber string
}

type ListPapersArgs struct{}

type ListPapersReply struct {
	Papers []struct {
		Number string
		Author string
		Title  string
	}
}

type GetPaperArgs struct {
	PaperNumber string
}

type GetPaperDetailsReply struct {
	Author string
	Title  string
}

type FetchPaperArgs struct {
	PaperNumber string
}

type FetchPaperReply struct {
	Content []byte
	Format  string
}
