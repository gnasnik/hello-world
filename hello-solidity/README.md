# Solidity 学习笔记

- [数据类型](#数据类型)
- [函数](#函数)
- [可视范围](#可视范围)
- [回退函数](#回退函数)
- [流程控制](#流程控制)
- [循环](#循环)
- [报错控制](#报错控制)
- [函数修改器](#函数修改器)
- [数组](#数组)
- [Map](#Map)
- [存储位置](#存储位置)
- [结构体](#结构体)
- [枚举](#枚举)
- [事件](#事件)
- [继承](#继承)
- [转账](#转账)
- 代理合约 - TODO


## 数据类型
- string
- bool
- int8/uint8, int256/uint256
- address/ address payable 可以使用 send 和 transfer
- bytes32 字节类型

```
contract DataType {
    string public name = "datatype";
    bool public isTest = true;
    uint8 public u = 1;
    uint256 public u256 = 256;

    int public maxInt = type(int).max;
    int public minInt = type(int).min;

    address public addr = 0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199;
    bytes32 public byte32 = "";

    // 使用 constant 常量也可以节省调用 gas
    address public constant ADDR_1 = 0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199;
    address public ADDR_2 = 0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199;

    // immutable 不可变量可以使变量成为常量，节省 gas 费, 也可以在构造函数里赋值
    address public immutable owner = msg.sender;

    // payable 修饰地址使这个地址可以接收eth或转账
    address payable public owner = payable(msg.sender);
}
```

address 有三以下成员:
- `balance`  查询地址余额
- `transfer` payable 地址可以转账， 失败会报错，并且 revert
- `send`     跟 transfer 一样，但失败不会报错，会返回 `false`

## 函数

- external 表名方法是对外公开的
- pure  纯计算，不读取合约里的变量
- view  只读合约里的变量，不修改

```
contract Function {
    uint256 public number = 256;

    // 读合约中的 number 变量要使用 view
    function add(uint256 x) external view returns (uint256) {
        return number + x;
    }

    // 没有读取合约的变量，用 pure 可以节省合约部署 gas 费
    function add(uint256 x, uint256 y) external pure returns (uint256) {
        return  x + y;
    }

    // 修改合约变量不能用有写操作，不能用 view
    function rewrite(uint256 x) external {
        number = x;
    }
}
```
## 可视范围
- `private` 只有内部可见 
- `internal` 内部和被继承的函数可见
- `public ` 外部和内部可见
- `external` 只有外部可见，只能由其他合约调用

## 回退函数
- `fallback` 当调用合约中一个不存在的函数或者调用空方法，亦或直接使用合约地址的内置函数 transfer() 或者 send() 时，都会执行目标合约的fallback 函数
- `receive`  如果只需要接收 Eth, 而不需要处理函数调用失败，用 receive 就够了

如果同时有`fallback` 和 `receive` 方法时, 没有 msg.data 调用 `receive` 方法，有 msg.data 的调用 `fallback`, 如果没有 `receive` 方法，那就会调用 `fallback`

```
contract MyCallback {
    event Log(string func, address caller, uint value, bytes data);

    fallback() external payable {
        emit Log("fallback", msg.sender, msg.value, msg.data);
    }

    receive() external payable {
        emit Log("receive", msg.sender, msg.value, "");
    }
}   
```

## 流程控制
- if-else 和 javescript 里的语法相似
```
if (x) < 10 {
    return 1;
}else if (x <20)  {
    return 2;
}else {
    return 3
}

```
三元运算符      
return x<10? true :false

## 循环
智能合约不能死循环

```
# for 
for (uint i = 0; i< 10; i++) {
    if (i == 3) {
        continue // 跳过
    }

    if (i == 4) {
        break // 跳出
    }
}


# while  
uint i = 0
while (i<10) {
    i++
}


# loop
contract Loop {
    function loop() external pure {
        for (uint i = 0; i< 100; i++ ) {
            // 
            if (i == 10) {
                continue;
            }

            if (i == 22) {
                break;
            }
        }


        // while
        uint j = 0; 
        while (j < 10) {
            j++;
        }

        //
    }
}
```

## 报错控制
- `require` require(表达式，错误信息); 状态会回滚，gas会返还
- `revert` revert(错误信息);
- `assert` assert(表达式); 可以用于测试

```
contract Error {
    function testRequire(uint i) public pure {
        require(i<10, "i>=10");
    }

    function testRevert(uint i) public pure {
        if (i < 10) {
            revert("i > 10);
        }
    }

    function testAssert(uint i) public pure{
        assert(i == 2);
    }
}
```

自定义错误, 好处是可以避免 `require` 的错误信息太长，造成gas浪费，节省gas
```
contract MyError{
    error TestError(address caller uint i);

    function testError(uint i) public view {
        if (i > 10) {
            revert TestError(msg.sender, i);
        }
    }
}
```

## 函数修改器

modifier 关键字来创建一个函数修改器，函数在调用前会先执行函数修改器

```
contract MyToken{
    uint32 public total;
    address public admin;

    modifier isAdmin() {
        require(msg.sender == admin, "sender no permission");
        _;
    }

    // 带参数
    modifier isAdmin2(address sender) {
        require(msg.sender == sender, "sender no permission");
        _;
    }

    // 函数前后执行修改器代码
    modifier isAdmin3()) {
        require(msg.sender == address(0), "sender no permission");
        _;
        require(msg.sender == sender, "sender no permission");
    }

    function addSupply(uint32 amount) external isAdmin {
        total += amount;
    }
}  
```

## 数组

- uint32[]  动态数组
- uint32[3] 定长数组

```
contract Array() {
    uint32[] nums;
    uint32[3] nums2;

    function arrayOp() external {
        nums.push(1); // 增加元素，定长数组不能 push
        delete nums[1];  // 把值设置为 0， 不会修改数组长度
        nums.pop(); // 取出最后一个元素
        nums.length;  // 获取数组长度
    }

    // 创建一个内存数组，必须指定长度，无法创建动态数组
    uint[] memory  a = new uint[](8);
    a[0] = 1;

    // 返回数组所有元素, 要加 memory
    function returnArray() external view returns (uint32[] memory) {
        return nums;
    }

}
```

## Map
```
contract Map() {
    mapping(address => uint) public balances;

    function addBalance() public external {
        balances[msg.sender] += 1;
    }

    function removeBalance() public external {
        delete balances[msg.sender]; // 恢复为默认值 0
    }
}
```

## 存储位置
 - `memory` 只存在内存中，执行完之后就回收,不能修改原值，相当于拷贝修改，参数传递是 string 和 数组时要用 memory 关键字
 - `storage` 可以修改原值，相当于引用修改
 - `calldata` 只能用在传递参数，跟 memory 相似，如果同个参数多个函数传递使用 calldata 可以节省 gas, 减少值的拷贝

```
contract DataLocation {
    struct MyStruct {
        uint foo;
        string text;
    }

    mapping(address => MyStruct) public myStructs;

    function example(uint[] memory y, string memory x) public external returns (uint[] memory) {
        myStructs[msg.sender] = MyStruct(123, "bar");

        MyStruct storage myStruct = myStructs[msg.sender];
        myStruct.text = "foo";

        uint[] memory arr = new uint[](3);
        //定长数组只能用索引添加值, 不能用 push
        arr[0] = 1;
        arr[1] = 2;

        return arr;
    }
}
```

## 结构体

```
contract MyStruct {
    struct Car {
        string model;
        uint year;
        address owner;
    }

    Car public car;
    Car[] public cars;
    mapping(address => Cars) public carsByOwner;

    function example() external {
        Car memory toyota = Car('Toyota', 1990, msg.sender);
        Car memory tesla = Car({model: 'Tesla', year: 2000, owner: msg.sender});

        cars.push(toyota);
        cars.push(tesla);

        // storage 可修改
        Car storage _car = cars[0];
        _car.year = 1980;
    }

}
```

## 枚举
```
contract MyEnums{
    enum Status {
        None,
        Pending,
        Shipped,
        Completed,
        Rejected,
        Canceled
    }

    Status public status;

    function get() public external view returns (Status) {
        return status;
    }

    function set(Status _status) external {
        status = _status;
    }

    function ship() external {
        status = Status.Shipped;
    }
}
```

## 事件
事件是记录当前智能合约运行状态的方法，通过事件可以查询我们修改过的状态。
event 定义一个事件， emit 触发事件， indexed 表示可以在索引的字段，可以在浏览器等筛选，一个事件里 indexed 字段不能超过3个

```
contract Event {
    event Log(string message, uint val);
    event IndexLog(address indexed sender, uint val);

    function printLog() external {
        emit Log('foo', 123);
        emit IndexLog(msg.sender, 812);
    }

    event Message(address indexed _from, address indexed _to, uint val);

    function sendMessage(address _to, string calldata message) external {
        emit Message(msg.sender, _to, message); 
    }
}
```

## 继承

继承可以避免重复写代码，使用 `is` 关键字, 多线继承要把简单的写在前面

```
contract Root{
    //  要使用 virtual 关键字表示方法可以被重写的
    function foo() public pure virtual returns (string memory) {
        return "Root";
    }

    function bar() public pure virtual returns (string memory) {
        return "Root";
    }

    function baz() public pure returns (string memory) {
        return "Root";
    }
}

contract Top is Root {
    // 使用 override 表示重写覆盖继承的方法
    function foo() public pure override returns (string memory) {
        return "Top";
    }

    function bar() public pure override returns (strings memory) {
        return "Top";
    }
}

```
## 转账
- `transfer` 2300 gas， 失败会 revert
- `send`  2300 gas, 返回是否成功
- `call` all gas, 返回是否成功和数据，有重入攻击风险

```
contract MyCallback {
    event Log(string func, address caller, uint value, bytes data);

    fallback() external payable {}
    receive() external payable {}
 
    function send() public payable {
        // 向当前合约转账
        // transfer 方法
        payable(this).transfer(msg.value);

        // send 方法
        bool suc = payable(this).send(msg.value);
        require(suc, "not success");
        
        // call 方法
        (bool success, ) = payable(this).call{value:msg.value}("");
        require(success, "not success");
    }

    function getBalance() public view returns (uint256) {
        // 获取当前合约地址
        return address(this).balance;
    }
}   
```

## 接口合约

使用 `interface` 来创建一个接口类，只要定义好方法名和参数返回值即可，调用时传入实际部署的合约地址。

```
interface ICounter {
    function count() external view returns (uint);
    function inc() external;
} 

contract CallInterface{
    uint public count;

    function example(address _counter) external {
        ICounter(_counter).inc();
        count = ICounter(_counter).count();
    }
}

```

## 低级调用call
```
contract TestCall {
    string public message;
    uint public x;

    fallback() external {}

    function foo(string memory _message, uint _x) external payable returns (bool, uint) {
        message = _message;
        x = _x;
        return (true, 666);
    }
}


contract Call {
    bytes public data;
    function foo(address _address) external payable {
        // 这里的 abi.encodeSignature 里面参数 uint 必须写成 256 位
        (bool suc, bytes memory _data) = _address.call{value: 100, gas: 21000}(abi.encodeWithSignature("foo(string, uint256)","invoke call", 6666));
        require(suc, "call failed");
        data = _data;
    }
}
```

## 委托调用合约
delegateCall 委托调用跟 call 写法一样，只不过委托调用不修改被调用的合约，而是更新调用者合约的值， 被调用合约的参数变量必须与调用者合约的参数变量对齐(也就是说参数顺序、类型必须一样)。

```
contract TestDelegateCall {
    address public sender;
    uint public x;
    uint public values;

    function setVars(uint _x) external payable {
        sender = msg.sender;
        values = msg.value;
        x = _x;
    }
}


contract DelegateCall {
    address public sender;
    uint public x;
    uint public values111;

    function delegateCall(address _test) external payable{
        (bool success, ) = _test.delegatecall(abi.encodeWithSignature("setVars(uint256)", 666));
        require(success, "call failed");
    }

    // 第二种调用的方法，适用 call, 不过好像 gas 会偏高？
    function delegateCall2(address _test) external payable {
        (bool success, ) = _test.delegatecall(abi.encodeWithSelector(TestDelegateCall.setVars.selector, 777));
        require(success, "call failed");
    }
}
```
## 工厂合约

```
contract Account{
    address public owner;
    address public bank;

    constructor(address _owner) payable {
        owner = _owner;
        bank = msg.sender;
    }
}


contract AccountFactory {
    Account[] public accounts;

    function createAccount(address _address) external payable {
        Account account = new Account{value:0}(_address);
        accounts.push(account);
    }
}
```
