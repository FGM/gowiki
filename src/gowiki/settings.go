package main

// ---- Settings ---------------------------------------------------------------

/*
Settings represent initial configuration values

	* These are initialized at program start from harcoded values.
	* Compare with Configuration.
*/
type Settings struct {
	Port             uint16
	PagesPath        string
	TemplatesPath    string
	Routes			 *RoutingMap
}



