# Makefile
# advices how to regenerate some .pngs
#
all: config
	@echo config file generated,
	@echo See Makefile for further hints

shoelaces-frontend-1.png:
	@echo
	@echo Information for after the \"sleep\"
	@echo Visit http://localhost:8081 with graphical webbrowser and screenshot program
	@echo Stop shoelaces binary with pressing Control-C
	@echo
	sleep 5
	@echo Starting the binary ...
	../../bin/shoelaces -config config

shoelaces-frontend-2.png:
	@echo
	@echo Undocumented
	@echo patches welcome
	@echo

shoelaces-overview.png: shoelaces-overview.msc
	msc-gen \
		-T png \
		-i $< \
		-o $@

config: config.in
	# because the `git clone` or `tar xf` could have been
	# done in any directory
	sed -e 's@$$PWD@${PWD}@' $< > $@

# FYI
#   msc-gen homepage is https://gitlab.com/msc-generator/msc-generator
#   screenshot program could be flameshot, https://flameshot.org/
#
# l l
