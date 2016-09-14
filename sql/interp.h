/*
 * interp.c: interpreter for RQL
 *
 * Authors: Dallan Quass (quass@cs.stanford.edu)
 *          Jan Jannink
 * originally by: Mark McAuliffe, University of Wisconsin - Madison, 1991
 */

#ifndef INTERP_H
#define INTERP_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/types.h>
#include "redbase.h"
#include "parser_internal.h"
#include "y.tab.h"


RC interp(NODE *n);
void QL_PrintError(RC rc);

#define QL_BADINSERT            (START_QL_WARN + 0) // Bad insert
#define QL_DUPRELATION          (START_QL_WARN + 1) // Duplicate relation
#define QL_BADSELECTATTR        (START_QL_WARN + 2) // Bad select attribute
#define QL_ATTRNOTFOUND         (START_QL_WARN + 3) // Attribute not found
#define QL_BADCOND              (START_QL_WARN + 4) // Bad condition
#define QL_BADCALL              (START_QL_WARN + 5) // Bad/invalid call
#define QL_CONDNOTMET           (START_QL_WARN + 6) // Condition has not been met
#define QL_BADUPDATE            (START_QL_WARN + 7) // Bad update statement
#define QL_EOI                  (START_QL_WARN + 8) // End of iterator
#define QO_BADCONDITION         (START_QL_WARN + 9)
#define QO_INVALIDBIT           (START_QL_WARN + 10)
#define QL_LASTWARN             QL_EOI

#define QL_INVALIDDB            (START_QL_ERR - 0)
#define QL_ERROR                (START_QL_ERR - 1) // error
#define QL_LASTERROR            QL_ERROR

#define PF_PAGEPINNED      (START_PF_WARN + 0) // page pinned in buffer
#define PF_PAGENOTINBUF    (START_PF_WARN + 1) // page isn't pinned in buffer
#define PF_INVALIDPAGE     (START_PF_WARN + 2) // invalid page number
#define PF_FILEOPEN        (START_PF_WARN + 3) // file is open
#define PF_CLOSEDFILE      (START_PF_WARN + 4) // file is closed
#define PF_PAGEFREE        (START_PF_WARN + 5) // page already free
#define PF_PAGEUNPINNED    (START_PF_WARN + 6) // page already unpinned
#define PF_EOF             (START_PF_WARN + 7) // end of file
#define PF_TOOSMALL        (START_PF_WARN + 8) // Resize buffer too small
#define PF_LASTWARN        PF_TOOSMALL

#define PF_NOMEM           (START_PF_ERR - 0)  // no memory
#define PF_NOBUF           (START_PF_ERR - 1)  // no buffer space
#define PF_INCOMPLETEREAD  (START_PF_ERR - 2)  // incomplete read from file
#define PF_INCOMPLETEWRITE (START_PF_ERR - 3)  // incomplete write to file
#define PF_HDRREAD         (START_PF_ERR - 4)  // incomplete read of header
#define PF_HDRWRITE        (START_PF_ERR - 5)  // incomplete write to header

// Internal errors
#define PF_PAGEINBUF       (START_PF_ERR - 6) // new page already in buffer
#define PF_HASHNOTFOUND    (START_PF_ERR - 7) // hash table entry not found
#define PF_HASHPAGEEXIST   (START_PF_ERR - 8) // page already in hash table
#define PF_INVALIDNAME     (START_PF_ERR - 9) // invalid PC file name

// Error in UNIX system call or library routine
#define PF_UNIX            (START_PF_ERR - 10) // Unix error
#define PF_LASTERROR       PF_UNIX


#endif

