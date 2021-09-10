# Phalcon 3 and Phalcon 4

## Phalcon 3

You can use [Phalcon+](https://github.com/bullsoft/phalconplus) to integrate with RR. Phalcon+ provides PSR-7 converter-classes: 
 - ```PhalconPlus\Http\NonPsrRequest``` to help converting PSR7 Request to Phalcon Native Request, 
 - ```PhalconPlus\Http\PsrResponseFactory``` to help create PSR7 Response from Phalcon Native Response.
 
and other finalizer to process stateful service in `di container`.

## Phalcon 4

Phalcon 4 has builtin supports for PSR-7:
 - [Request](https://docs.phalcon.io/4.0/zh-cn/http-request),
 - [Response](https://docs.phalcon.io/4.0/zh-cn/http-response), 
 
 you can easily integrate with RR.
