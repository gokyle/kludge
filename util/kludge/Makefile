TARGET = kludge
SOURCES = client.go
INSTALL_PATH = /usr/local/bin

all: $(TARGET)

$(TARGET): $(SOURCES)
	go build -o $(TARGET)

install: $(TARGET)
	install $(TARGET) $(INSTALL_PATH)/$(TARGET)

clean:
	rm -f $(TARGET)

uninstall:
	rm $(INSTALL_PATH)/$(TARGET)

.PHONY: all clean install uninstall
