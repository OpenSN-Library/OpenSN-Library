#include <stdio.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <string.h>
#include <linux/netlink.h>
#include <stdint.h>
#include <unistd.h>
#include <errno.h>

#define NETLINK_TEST    30
#define MSG_LEN            200
#define MAX_PLOAD        400

typedef struct _user_msg_info
{
    struct nlmsghdr hdr;
    char  msg[MSG_LEN];
} user_msg_info;

void readLineWithoutNewline(char*s, int n, FILE*stream)
{
    fgets(s, n, stream);
    if (s[strlen(s)-1] == '\n')
    {
        s[strlen(s)-1] = '\0';
    }
}

int main(int argc, char **argv)
{
    int skfd;
    int ret;

    socklen_t len;
    struct nlmsghdr *nlh = NULL;
    struct sockaddr_nl saddr, daddr; //saddr 表示源端口地址，daddr表示目的端口地址
    char *umsg = NULL;

    skfd = socket(AF_NETLINK, SOCK_RAW, NETLINK_TEST);
    int sndbuf = 1024 * 1024 * 8;
    int err = setsockopt(skfd, SOL_SOCKET, SO_SNDBUF, &sndbuf, sizeof(sndbuf));
    err = setsockopt(skfd, SOL_SOCKET, SO_RCVBUF, &sndbuf, sizeof(sndbuf));
    if(err < 0){
        printf("setsockopt error: %s\n", strerror(errno));
        return -1;
    }

    if(skfd == -1)
    {
        perror("create socket error\n");
        return -1;
    }

    memset(&saddr, 0, sizeof(saddr));
    saddr.nl_family = AF_NETLINK; //AF_NETLINK
    saddr.nl_pid = 100;  //端口号(port ID)
    saddr.nl_groups = 0;
    if(bind(skfd, (struct sockaddr *)&saddr, sizeof(saddr)) != 0)
    {
        perror("bind() error\n");
        close(skfd);
        return -1;
    }

    memset(&daddr, 0, sizeof(daddr));
    daddr.nl_family = AF_NETLINK;
    daddr.nl_pid = 0; // to kernel
    daddr.nl_groups = 0;
    // use while loop
    while(1){
        user_msg_info u_info;
        nlh = (struct nlmsghdr *)malloc(NLMSG_SPACE(MAX_PLOAD));
        memset(nlh, 0, sizeof(struct nlmsghdr));
        nlh->nlmsg_len = NLMSG_SPACE(MAX_PLOAD);
        nlh->nlmsg_flags = 0;
        nlh->nlmsg_type = 0;
        nlh->nlmsg_seq = 0;
        nlh->nlmsg_pid = saddr.nl_pid; //self port
        // user input
        umsg = (char *)malloc(MAX_PLOAD);
        memset(umsg, 0, MAX_PLOAD);
        printf("please input message(quit or q to exit):");
        // get a line input
        readLineWithoutNewline(umsg, MAX_PLOAD, stdin);
        if (strlen(umsg) == 0){
            printf("input message is empty, please input again\n");
            continue;
        }
        if (strcmp(umsg, "quit") == 0 || strcmp(umsg, "q") == 0)
        {
            printf("quit\n");
            break;
        }
        memcpy(NLMSG_DATA(nlh), umsg, strlen(umsg)+1); // 要拷贝空字符
        ret = sendto(skfd, nlh, nlh->nlmsg_len, 0, (struct sockaddr *)&daddr, sizeof(struct sockaddr_nl));
        if(!ret)
        {
            perror("sendto error\n");
            close(skfd);
            exit(-1);
        }
        printf("send kernel:%s\n", umsg);

        memset(&u_info, 0, sizeof(u_info));
        len = sizeof(struct sockaddr_nl);
        while(1){
            ret = recvfrom(skfd, &u_info, sizeof(user_msg_info), 0, (struct sockaddr *)&daddr, &len);
            if (strcmp(u_info.msg, "") == 0)
            {
                continue;
            } else {
                break;
            }
        }
        if(!ret)
        {
            perror("recv form kernel error\n");
            close(skfd);
            exit(-1);
        }

        printf("from kernel:%s\n", u_info.msg);

        free(nlh);

    }
    free(umsg);
    free(nlh);
    close(skfd);

    return 0;
}
