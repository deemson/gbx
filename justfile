yo:
	echo {{justfile_directory()}}
	echo {{invocation_directory()}}

simulate name:
	GBX_SIMULATION_NAME={{name}} GBX_SIMULATIONS_OUTPUT={{justfile_directory()}}/simulations go test {{justfile_directory()}}/internal/git/simulate -v
