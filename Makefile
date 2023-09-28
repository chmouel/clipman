OUTPUT_DIR = bin
NAME  := clipman

all: $(OUTPUT_DIR)/$(NAME)

mkdir: $(OUTPUT_DIR)
	mkdir -p $(OUTPUT_DIR)

$(OUTPUT_DIR)/$(NAME): *.go mkdir
	go build $(FLAGS)  -v -o $@ ./
