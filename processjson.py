# /usr/local/bin/python3
# -*- coding: UTF-8 -*-
# auth: 大帅逼
# funciton: deal the json data
import getopt
import inspect
import json
import os
import subprocess
import sys
from sys import argv

_delay = ""
text = ""
slowlogfile = ""
starttime = ""
endtime = ""
_reportfile = "report.md"


## 将无规律的json解析成我要环境
## 后面计划直接 cat mongod的文件--> 管道,然后遍历每一行然后做处理，进行简单的问题分析


def scriptname():
    return inspect.getfile(sys.modules[__name__])


def str2json(a):
    global data
    _output = {}  # init dict

    try:
        data = json.loads(a)
    except:
        print("decode json for text '{}' err ,exit ...".format(a))
        exit(1)

    # attr 下的
    _attr = data["attr"]
    keys_attr = ["ns", "durationMillis", "planSummary", "docsExamined", "storage", "appName"]

    # attr.command下的
    _addr_common = data["attr"]["command"]
    keys_command = ["insert", "update", "q", "u"]

    if (dicthaskey(_attr, "durationMillis")) and (_attr["durationMillis"]) >= 100:
        _output["time"] = data["t"]["$date"]
        for i in keys_command:

            if dicthaskey(_addr_common, i):
                _output[i] = _addr_common[i]

        for i in keys_attr:

            if dicthaskey(_attr, i):
                _output[i] = _attr[i]

        return str(_output)


def dicthaskey(dict, key):
    if key in dict.keys():
        return True
    else:
        return False


def AinB(stra, strb):
    if stra in strb:
        return True
    else:
        return False


def init():
    script_name = scriptname()
    argv = sys.argv[1:]
    global _delay, text, slowlogfile, data

    try:
        opts, args = getopt.getopt(argv, "h:i:f:s:",
                                   ["input=", "file=", "slow="])
    except getopt.GetoptError:
        print(script_name, ' -i <text> -f <file> -d <慢日志阈值>')
        sys.exit(2)

    for opt, arg in opts:
        if opt == '-h':
            print(script_name, ' -i <text>')
            sys.exit(2)
        elif opt in ("-i", "--input"):
            text = arg
        elif opt in ("-f", "--file"):
            slowlogfile = arg
        elif opt in ("s", "--slow"):
            _delay = int(arg)

    if _delay is None:
        _delay = 10

    if text != "":
        str2json(text)

    # if slowlogfile != "":
    #     job(slowlogfile)


def runCmd(cmd):
    command = cmd
    subprocess.getoutput(command)


def getoutput(cmd):
    # 执行cmd命令，如果成功，返回(0, 'xxx')；如果失败，返回(1, 'xxx')
    res = subprocess.Popen(cmd, shell=True, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                           stderr=subprocess.PIPE)  # 使用管道
    result = res.stdout.read()  # 获取输出结果
    res.wait()  # 等待命令执行完成
    res.stdout.close()  # 关闭标准输出
    return result


def adb_shell(cmd):
    result = os.popen(cmd).read()
    return result


def initreportfile(file):
    ## 初始化文件
    with open(file, "w") as fd:
        fd.close()


def writemd(file, header, info, text):
    with open(file, "a") as fd:
        fd.write("## ")
        fd.write(header + "\n")
        fd.write(info + "\n")
        fd.write("```bash\n")
        fd.write(text + "\n")
        fd.write("```\n")
        fd.close()


class analylogmessage:
    def __init__(self, body, f1, f2):
        self.start = True
        self.end = True
        self.slowlogname = body
        self.session1 = False
        self.session2 = False
        if f1 == "":
            self.start = False
        if f2 == "":
            self.end = False
        if self.start and self.end:
            # 既有starttime 又有 endtime的场景
            self.session1 = True
        elif not self.start and not self.end:
            # 不加时间的场景
            self.session2 = True
        # else:
        # 只有starttime参数

    def analysizeclientkind(self):
        if self.session1:
            command = "jq '. | select(.t[\"$date\"] >= \"" + starttime + "\" and .t[\"$date\"] <= \"" + endtime + "\")' " + self.slowlogname + " |jq -r '.attr.doc.os.type' " + "| grep -v null | sort | uniq -c | sort -rn"
        elif self.session2:
            command = "jq -r '.attr.doc.os.type' " + self.slowlogname + "| grep -v null | sort | uniq -c | sort -rn"
        else:
            command = "jq '. | select(.t[\"$date\"] >= \"" + starttime + "\" )' " + self.slowlogname + " |jq -r '.attr.doc.os.type' " + "| grep -v null | sort | uniq -c | sort -rn"

        header = "客户端类型分析"
        info = "以下示例分析报告的远程[MongoDB 驱动程序](https://docs.mongodb.com/ecosystem/drivers/) 连接和客户端应用程序的[客户端数据](https://docs.mongodb.com/manual/reference/log-messages/#std-label-log-messages-client-data)，包括，并打印连接的每个唯一操作系统类型的总数，按频率排序"
        # write report file
        msg = adb_shell(command)
        print(header, "\n ", info)
        print(str(msg))
        writemd(_reportfile, header, info, msg)

    def analydriverconn(self):
        # 以下示例计算所有远程[MongoDB 驱动程序](https://docs.mongodb.com/ecosystem/drivers/)连接数，并按出现次数降序显示每个驱动程序类型和版本
        # jq -cr '.attr.doc.driver' /var/log/mongodb/mongod.log | grep -v null | sort | uniq -c | sort -rn
        # ```
        #  14 {"name":"mongo-go-driver","version":"v1.7.1"}
        #  3 {"name":"NetworkInterfaceTL","version":"4.4.6"}
        #  3 {"name":"MongoDB Internal Client","version":"4.4.6"}
        #  ```

        if self.session1:
            command = "jq '. | select(.t[\"$date\"] >= \"" + starttime + "\" and .t[\"$date\"] <= \"" + endtime + "\")' " + self.slowlogname + " | jq -cr '.attr.doc.driver' | grep -v null | sort | uniq -c | sort -rn"
        elif self.session2:
            command = "jq -cr '.attr.doc.driver' " + self.slowlogname + "| grep -v null | sort | uniq -c | sort -rn"
        else:
            command = "jq '. | select(.t[\"$date\"] >= \"" + starttime + "\" )' " + starttime + " | jq -cr '.attr.doc.driver' | grep -v null | sort | uniq -c | sort -rn"

        # write report file
        header = "Application MongoDB驱动程序分析"
        info = "以下为计算得到的所有远程[MongoDB 驱动程序](https://docs.mongodb.com/ecosystem/drivers/)连接数，并按出现次数降序显示每个驱动程序类型和版本"
        msg = adb_shell(command)
        print(header, "\n ", info)
        print(str(msg))
        writemd(_reportfile, header, info, msg)

    def analyconnremote(self):
        # 远程客户端连接显示在日志中属性对象的“remote” key下。
        if self.session1:
            command = "jq '. | select(.t[\"$date\"] >= \"" + starttime + "\" and .t[\"$date\"] <= \"" + endtime + "\")' " + self.slowlogname + "|jq -r '.attr.remote'| grep -v 'null' | awk -F':' '{print $1}' | sort | uniq -c | sort -r"
        elif self.session2:
            command = "jq -r '.attr.remote' " + self.slowlogname + " |grep -v 'null' | awk -F':' '{print $1}' | sort | uniq -c | sort -r"
        else:
            command = "jq '. | select(.t[\"$date\"] >= \"" + starttime + "\" )' " + self.slowlogname + " |jq -r '.attr.remote'| grep -v 'null' | awk -F':' '{print $1}' | sort | uniq -c | sort -r"

        header = "Application Client分析"
        info = "以下为计算得到的应用程序客户端的连接情况统计，并按出现次数降序显示每个驱动程序的类型和版本"
        msg = adb_shell(command)
        print(header, "\n ", info)
        print(str(msg))
        writemd(_reportfile, header, info, msg)

    def unionmessage(self):
        # 以下示例显示给定日志文件中的前 10 个唯一消息值，按频率排序：
        if self.session1:
            command = "jq '. | select(.t[\"$date\"] >= \"" + starttime + "\" and .t[\"$date\"] <= \"" + endtime + "\")' " + self.slowlogname + " | jq -r '.msg'| sort | uniq -c | sort -rn | head -10"
        elif self.session2:
            command = "jq -r '.msg' " + self.slowlogname + " |sort | uniq -c | sort -rn | head -10 "
        else:
            command = "jq '. | select(.t[\"$date\"] >= \"" + starttime + "\" )' " + self.slowlogname + " | jq -r '.msg'| sort | uniq -c | sort -rn | head -10"
        header = "MongoDB log message 唯一消息统计"
        info = "通过分析mongodb日志的唯一消息，可以直观的分析mongodb log message的信息"
        msg = adb_shell(command)
        print(header, "\n ", info)
        print(str(msg))
        writemd(_reportfile, header, info, msg)

    def anlayslowlogjob(self):
        _slowlogflag = "Slow query"
        _slowoplog = "\"ns\":\"local.oplog.rs\""
        cmd = "cat " + self.slowlogname
        result = os.popen(cmd, mode='r')
        header = "慢查询分析"
        info = "慢查询记录了执行缓慢的语句，通过分析慢查询的规律来找到最需要优化的内容"
        slowloglist = []
        for line in result.read().split('\n'):
            if AinB(_slowlogflag, line) and not AinB(_slowoplog, line):
                slowlog = str2json(line)
                if slowlog != None:
                    slowloglist.append(slowlog)
        text=""
        for i in slowloglist:
            text=text+i+"\n"
        print(header, "\n ", info)
        print(text)
        writemd(_reportfile, header, info, text)


def main():
    ##  这部分是MongoDB的日志的巡检  ##
    init()
    initreportfile(_reportfile)
    analy = analylogmessage(slowlogfile, starttime, endtime)
    analy.unionmessage()
    analy.analysizeclientkind()
    analy.analydriverconn()
    analy.analyconnremote()
    analy.anlayslowlogjob()
    ## 待完成事项 ##
    ## 1、MongoDB数据库指标的采集
    ## 2、MongoDB数据库所在主机关键系统参数采集(此参数采集来自于自建的mongodb)
    ## 3、Mongo shell里面的那些提示的，不清楚如何得到的那些Warning，比如 内核参数文件，numa，文件打开数等等

if __name__ == '__main__':
    main()
