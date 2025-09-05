import signal
import time
from pid import PidFile

# 服务运行状态，为 True 就需要主动释放一些资源了
exit_flag = False

def stop_handler(sig, frame) :
    global exit_flag
    # 设置进程为待关闭状态
    exit_flag = True
    print("get "+signal.strsignal(sig)+" "+str(sig), flush=True)


# 启动自己的服务
def run_server() :
    while True:
        # 判断是否接收到中断服务的信号
        if exit_flag :
            ## 等待未完成的服务处理完
            ## 回收资源，关闭文件描述符等等
            break
        ## 在这里运行自己需要的服务，申请资源等等
        print("still running", flush=True)
        time.sleep(1)
        pass

if __name__ == "__main__" :
    # 响应 SIGINT 2 中断信号
    signal.signal( signal.SIGINT, stop_handler )
    # 响应 SIGTERM 15 请求进程终止信号
    signal.signal( signal.SIGTERM, stop_handler )
    with PidFile(pidname="tsys.pid", piddir=".") :
        run_server()
