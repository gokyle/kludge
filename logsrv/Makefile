TARGET = logsrv
SOURCES = server.go
INSTALL_DIR = /usr/local
BINDIR = $(INSTALL_DIR)/bin
MANDIR = $(INSTALL_DIR)/share/man/man1
MANPAGE = $(MANDIR)/$(TARGET).1
INITDIR = /etc/init.d/
INITSRC = etc/init.d
DBDIR = /var/run/$(TARGET)

all: $(TARGET)

$(TARGET): $(SOURCES)
	go build -o $(TARGET)

$(MANDIR):
	install -d $(MANDIR)

$(MANPAGE): $(MANDIR)
	install $(TARGET).1 $(MANPAGE)

$(INITDIR)/$(TARGET): $(INITSRC)/$(TARET)
	install -m 755 $(INITSRC)/$(TARGET) $(INITDIR)/$(TARGET)

$(DBDIR):
	install -d $(DBDIR)

install: $(TARGET) $(MANPAGE) $(INITDIR)/$(TARGET) $(DBDIR)
	install $(TARGET) $(BINDIR)/$(TARGET)

uninstall:
	rm -f $(MANPAGE)
	rm -f $(INITDIR)/$(TARGET)
	rm -f $(BINDIR)/$(TARGET)

clean:
	rm -f $(TARGET)

.PHONY: all clean install uninstall
