原项目地址 https://github.com/davyxu/tabtoy
基于 2021.6.15 版本修改
原版本commit 7649be588db79484c4ee2ef0ea800f17ba5fc4da


测试运行时v2时使用命令行参数如下:
--mode=v2 --csharp_out=.\csharp_outputs\Config.cs --binary_out=.\binary_outputs\ConfigBin.bytes --lua_out=.\lua_outputs\ConfigLua.lua --cpp_out=.\cpp_outputs\Config.hpp --go_out=.\go_outputs\Config.go --json_out=.\json_outputs\Config.json --combinename=Config --lan=zh_cn globals.xlsx command.xlsx


C# Unity 加载配置表代码示例:
# 默认把ConfigBin.bytes 文件放到 Resources目录下(如果是Resources下的子目录, 需要给出相对路径)
# configTable.GetBuildID() 函数返回的是 Config.cs 文件中写死的 buildId, 如果要考虑配置表热更新, 这个要换成从服务器获取最新的 buildId
void Start()
    {
        string path = "ConfigBin";
        var bytes = Resources.Load<TextAsset>(path);
        Debug.Log("bytes.size = " + bytes.bytes.Length);
        using (MemoryStream stream = new MemoryStream(bytes.bytes))
        {
            stream.Position = 0;
            var reader = new DataReader(stream);
            configTable = new table.Config();
            
            var result = reader.ReadHeader(configTable.GetBuildID());
            if (result != FileState.OK)
            {
                Debug.Log("combine file crack! " + result.ToString());
                return;
            }

            table.Config.Deserialize(configTable, reader);
        }
    }
	
使用说明:
1. 必须包含 globals.xlsx, 其中必须包含 @Types 标签页, 记录所有类型信息
2. 在 globals.xlsx 可以选择性的包含一个 @OutputTag 标签页, 用来标记特定的表格输出到特定的语言中。如果没有, 则表示所有表格输出到所有语言中。特别注意: 这个标签页中如果出现第一列为空的行, 就会被视为空行, 读到空行就会终止读取; 所以务必包正连续填写内容不要空行