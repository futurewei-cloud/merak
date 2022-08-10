module := services

submodules := proto scenario-manager
-include $(patsubst %, $(module)/%/module.mk, $(submodules))

all:: $(submodules)
