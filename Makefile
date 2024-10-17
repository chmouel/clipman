OUTPUT_DIR = bin
NAME  := clipman
GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)
GOFUMPT := $(shell command -v gofumpt 2> /dev/null)

all: lint $(OUTPUT_DIR)/$(NAME)

mkdir: $(OUTPUT_DIR)
	@mkdir -p $(OUTPUT_DIR)

$(OUTPUT_DIR)/$(NAME): *.go mkdir
	@echo "building..."
	@go build $(FLAGS)  -v -o $@ ./

lint: $(GOLANGCI_LINT)
	@echo "linting..."
	@$(GOLANGCI_LINT) run

fumpt:
	@find . -name '*.go'|xargs -P4 $(GOFUMPT) -w -extra


.PHONY: fumpt lint mkdir all
