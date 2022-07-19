module := services

submodules := merak-compute scenario-manager merak-topo
-include $(patsubst %, $(module)/%/module.mk, $(submodules))

all:: $(submodules)
