
<a href="https://github.com/HeavyHorst/roachbalancer"><img src="https://img.shields.io/badge/api-reference-blue.svg?style=flat-square" alt="GoDoc"></a>


# roachbalancer
roachbalancer is a small experimental cockroachdb load balancer with automatic live-node discovery.
The loadbalancer can be embedded or run as a stand-alone application.

## Usage
 ### Embedded
 ```go
func  main()  {
	b  :=  balancer.New("root",  "/certs",  false,  "xxx.xxx.xxx.xxx:26257", "xxx.xxx.xxx.xxx:26257")
	go  b.Listen(0)  //  0  means  random  high  port
	b.WaitReady()

	// now use b.GetAddr() as the cockroachdb address
}
 ```

### Use the binary

```
./roachbalancer -certs-dir /certs -node xxx.xxx.xxx.xxx:26257
```