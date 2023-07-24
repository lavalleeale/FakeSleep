// C++ Program to print current Day, Date and Time
#include <iostream>
#include <unistd.h>
#include <sys/syscall.h>
using namespace std;
int main()
{
    printf("Modified Time Response: %d\n", syscall(SYS_time));
    sleep(4);
    printf("Modified Time Response: %d\n", syscall(SYS_time));

    return 0;
}