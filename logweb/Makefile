TARGET = logweb
SOURCES = http.go db.go
INSTALL_DIR = /usr/local
BINDIR = $(INSTALL_DIR)/bin
MANDIR = $(INSTALL_DIR)/share/man/man1
MANPAGE = $(MANDIR)/$(TARGET).1
INITDIR = /etc/init.d/
INITSRC = etc/init.d
DBDIR = /var/run/$(TARGET)
SHARE = /usr/local/share
CONFIG = logwebrc.master
CONFDIR = /etc/kludge
TEMPLATES = templates/index.html templates/logs.html templates/response.html

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

$(SHARE):
	install -d $(SHARE)

$(CONFDIR):
	install -d $(CONFDIR)

$(CONFDIR)/$(TARGET)rc: $(CONFIG)
	install -m 0644 $(CONFIG) $(CONFDIR)/$(TARGET)rc

assets-install: assets/css/bootstrap.css
	install -C -m 0644 assets $(SHARE)/$(TARGET)

templates-install: $(TEMPLATES)
	install -C -d $(SHARE)/$(TARGET)/templates
	install -C -m 0644 $(TEMPLATES) $(SHARE)/$(TARGET)/templates/

install: $(TARGET) $(MANPAGE) $(INITDIR)/$(TARGET) $(DBDIR)
	install $(TARGET) $(BINDIR)/$(TARGET)

uninstall:
	rm -f $(MANPAGE)
	rm -f $(INITDIR)/$(TARGET)
	rm -f $(BINDIR)/$(TARGET)

clean:
	rm -f $(TARGET)

.PHONY: all clean install uninstall
