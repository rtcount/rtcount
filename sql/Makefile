#
# Makefile for the RedBase project.
#
# Author: Hyunjung Park (hyunjung@cs.stanford.edu)
#


#
# Global variables
#
CC             = g++
BUILD_DIR      = ./build/
LIB_DIR        = ./lib/
INC_DIRS       = -I.
AR             = ar -rc
RANLIB         = ranlib
YACC           = bison -dy
LEX            = flex

# -m32 - Generates code that runs on any i386 system
#        sizeof(int) = sizeof(long) = sizeof(void *) = 4
# -g - Debugging information
# -O1 - Basic optimization
# -Wall - All warnings
# -DDEBUG_PF - This turns on the LOG file for lots of BufferMgr info
CFLAGS         = -m32 -g -O1 -Wall $(STATS_OPTION) $(INC_DIRS)

# The STATS_OPTION can be set to -DPF_STATS or to nothing to turn on and
# off buffer manager statistics.  The student should not modify this
# flag at all!
STATS_OPTION   = -DPF_STATS

#
# Students: Please modify SOURCES variables as needed.
#
PARSER_SOURCES = scan.c parse.c nodes.c 
UTILS_SOURCES  = interp.cc

UTILS_OBJECTS  = $(addprefix $(BUILD_DIR), $(UTILS_SOURCES:.cc=.o))
PARSER_OBJECTS = $(addprefix $(BUILD_DIR), $(PARSER_SOURCES:.c=.o))
OBJECTS        = $(QL_OBJECTS) $(PARSER_OBJECTS) $(UTILS_OBJECTS)

LIBRARY_PARSER = $(LIB_DIR)libparser.a
LIBRARIES      = $(LIBRARY_PARSER)

UTILS          = $(UTILS_SOURCES:.cc=)
EXECUTABLES    = $(UTILS)

LIBS           = -lparser

#
# Build targets
#
all: $(LIBRARIES)

clean:
	rm -f $(BUILD_DIR)*.o $(BUILD_DIR)*.d y.output y.tab.h scan.c parse.c $(LIBRARIES) $(EXECUTABLES)

testers: all $(TESTS)

#
# Libraries
#

$(LIBRARY_PARSER): $(PARSER_OBJECTS)
	$(AR) $(LIBRARY_PARSER) $(PARSER_OBJECTS)
	$(RANLIB) $(LIBRARY_PARSER)
	./b.sh	

#
# Parser
#
y.tab.h: parse.c

parse.c: parse.y
	$(YACC) parse.y; mv y.tab.c parse.c

scan.c: scan.l scanhelp.c y.tab.h
	$(LEX) scan.l; mv lex.yy.c scan.c

$(BUILD_DIR)parse.o: parse.c

$(BUILD_DIR)scan.o: scan.c y.tab.h

$(BUILD_DIR)nodes.o: nodes.c

$(BUILD_DIR)interp.o: interp.cc

#
# Rules
#
-include $(OBJECTS:.o=.d)

$(BUILD_DIR)%.d: %.cc
	@set -e; \
	 rm -f $@; \
	 $(CC) $(CFLAGS) -MM -MT $(@:.d=.o) $< > $@.$$$$; \
	 sed 's,\($*\)\.o[ :]*,\1.o $@: ,g' $@.$$$$ > $@; \
	 rm -f $@.$$$$

$(OBJECTS): %.o:
	$(CC) $(CFLAGS) -c $< -o $@

$(EXECUTABLES): %: $(BUILD_DIR)%.o $(LIBRARIES)
	$(CC) $(CFLAGS) $< -o $@ -L$(LIB_DIR) $(LIBS)

