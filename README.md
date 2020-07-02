# mpc
## 说明  
安全多方计算demo（三个参与方和一个中央服务器），可以根据实际情况调整其他算法

## 运行  
修改配置文件config.yaml  
服务端运行  
./secureServer  

参与方一运行  
./secureClient 1  

参与方二运行  
./secureClient 2 

参与方三运行  
./secureClient 3  

## 开始测试  
服务端控制台输入启动命令:1  
最终服务端会统计三个参与方生成的随机数总和，但是服务端无法反推出每个参与方的随机数具体值  
