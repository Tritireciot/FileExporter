package html

import (
	"PrintServer/html2docx/utils"
)

func (tag Tag) IsContainer() bool {
	return utils.Contains(ContainerTags, tag)
}

func (tag Tag) IsList() bool {
	return utils.Contains(ListTags, tag)
}

func (tag Tag) IsTable() bool {
	return utils.Contains(TableTags, tag)
}
