package actor

func (c *Context) SetMetaData(meta *Meta) {
	c.MetaData = meta
}

func (c *Context) SetProperties(properties map[string]string) {
	c.Properties = properties
}
