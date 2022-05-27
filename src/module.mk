module := src

submodules := merak-compute
-include $(patsubst %, $(module)/%/module.mk, $(submodules))

all:: $(submodules)
