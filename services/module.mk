module := services

submodules := merak-compute scenario-manager
-include $(patsubst %, $(module)/%/module.mk, $(submodules))

all:: $(submodules)
