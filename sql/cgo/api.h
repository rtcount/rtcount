
#ifndef __API_H__
#define __API_H__

#ifdef __cplusplus    //__cplusplus是cpp中自定义的一个宏
extern "C" {          //
#endif


extern void hello();
extern int Num;

extern const char* ddd(const char * str);

#ifdef __cplusplus
}
#endif
#endif
