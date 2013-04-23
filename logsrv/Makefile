TARGET = logsrv
SOURCES = server.go
INSTALL_DIR = /usr/local
BINDIR = $(INSTALL_DIR)/bin
MANDIR = $(INSTALL_DIR)/share/man/man1
MANPAGE = $(MANDIR)/$(TARGET).1
UPSTART_DIR = /etc/init
UPSTART = $(TARGET).conf

all: $(TARGET)

$(TARGET): $(SOURCES)
	go build -o $(TARGET)

$(MANDIR):
	install -D $(MANDIR)

$(MANPAGE): $(MANDIR)
	install $(TARGET).1 $(MANPAGE)

$(UPSTART_DIR)/$(UPSTART): $(UPSTART)
	install $(UPSTART) $(UPSTART_DIR)/$(UPSTART)

install: $(TARGET) $(MANPAGE) $(UPSTART_DIR)/$(UPSTART)
	install $(TARGET) $(BINDIR)/$(TARGET)

uninstall:
	rm -f $(MANPAGE)
	rm -f $(UPSTART_DIR)/$(UPSTART)
	rm -f $(BINDIR)/$(TARGET)

clean:
	rm -f $(TARGET)

.PHONY: all clean install uninstall