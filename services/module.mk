module := services

submodules := proto merak-compute scenario-manager merak-agent
-include $(patsubst %, $(module)/%/module.mk, $(submodules))

all:: $(submodules)
