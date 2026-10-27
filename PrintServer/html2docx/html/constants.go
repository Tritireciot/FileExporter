package html

// Tags

type Tag string

const (
	StyleTag      Tag = "style"
	BodyTag       Tag = "body"
	TableTag      Tag = "table"
	Header1Tag    Tag = "h1"
	Header2Tag    Tag = "h2"
	Header3Tag    Tag = "h3"
	Header4Tag    Tag = "h4"
	Header5Tag    Tag = "h5"
	Header6Tag    Tag = "h6"
	LinkTag       Tag = "a"
	DivTag        Tag = "div"
	SummaryTag    Tag = "summary"
	DetailsTag    Tag = "details"
	SectionTag    Tag = "section"
	HeaderTag     Tag = "header"
	FooterTag     Tag = "footer"
	DLTag         Tag = "dl"
	ULTag         Tag = "ul"
	OLTag         Tag = "ol"
	MenuTag       Tag = "menu"
	TheadTag      Tag = "thead"
	TbodyTag      Tag = "tbody"
	TfootTag      Tag = "tfoot"
	TrTag         Tag = "tr"
	ThTag         Tag = "th"
	TdTag         Tag = "td"
	StrongTag     Tag = "strong"
	BTag          Tag = "b"
	EmTag         Tag = "em"
	ITag          Tag = "i"
	UTag          Tag = "u"
	InsTag        Tag = "ins"
	DelTag        Tag = "del"
	STag          Tag = "s"
	SmallTag      Tag = "small"
	CodeTag       Tag = "code"
	BrTag         Tag = "br"
	DDTag         Tag = "dd"
	BlockQuoteTag Tag = "blockquote"
	HrTag         Tag = "hr"
	ImgTag        Tag = "img"
	SubTag        Tag = "sub"
	SupTag        Tag = "sup"
)

var ContainerTags []Tag = []Tag{
	DivTag, SummaryTag, DetailsTag, SectionTag, HeaderTag, FooterTag, DLTag,
}

var ListTags []Tag = []Tag{
	ULTag, OLTag, MenuTag,
}

var TableTags []Tag = []Tag{
	TheadTag, TbodyTag, TfootTag,
}

// Attributes

type HTMLAttr string

const (
	SourceAttr  HTMLAttr = "src"
	HREFAttr    HTMLAttr = "href"
	StyleAttr   HTMLAttr = "style"
	RowSpanAttr HTMLAttr = "rowspan"
	ColSpanAttr HTMLAttr = "colspan"
	ClassAttr   HTMLAttr = "class"
	IdAttr      HTMLAttr = "id"
)
