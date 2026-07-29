package css

type StyleMap map[StyleProperty]StyleValue

func (styleMap StyleMap) Update(otherMap StyleMap) {
	for property, value := range otherMap {
		styleMap[property] = value
	}
}
