
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#ifdef __cplusplus    //__cplusplus是cpp中自定义的一个宏
extern "C" {          //告诉编译器，这部分代码按C语言的格式进行编译，而不是C++的
#endif
const char*  RBparse(const char * str);

const char* ddd(const char * str);

#ifdef __cplusplus
}
#endif

char  xml_parser[1024*10];
//int ddd(const char* str)
const char* ddd(const char * str)
{
	char * dd = (char*)RBparse(str);
	if (dd==NULL)
		return "";

	memset(xml_parser, 0, 1024*10);
	memcpy(xml_parser, dd, strlen(dd));
	free(dd);
	return (const char*)xml_parser;
/*
	   char tstr[] = "select asdddd from T_devices.ddd with zxc and xxx where created_at >= 1466252795 and product_id = 'T_devices.product_id' and product_id = \"T_devices.product_id\" and time =\"ad\" and time >'asd';";

	   RBparse(tstr);

	   char tstr2[] = "select ddd from ddd.ddd with ddd and ddd where ddd >= 1466252795 and ddd_id = 'ddd' and ddd = \"ddd\" and ddd =\"ddd\" and ddd >'ddd';";

	   RBparse(tstr2);
*/
}

/*
int main()
{
return ddd();
}
*/
