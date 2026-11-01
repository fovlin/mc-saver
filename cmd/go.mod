module acovia.net/mc-saver

go 1.27.2

replace acovia.net/record => ../record

replace acovia.net/opt => ../opt

replace acovia.net/minecraft/save => ../save

require (
	acovia.net/minecraft/save v0.0.0
	acovia.net/opt v0.0.0
	acovia.net/record v0.0.0
	golang.org/x/term v0.47.0
)

require golang.org/x/sys v0.49.0 // indirect
