```
             Internet
                |
           NAT Gateway
                |
           Security VPC
        +------------------+
        | Gateway LB       |
        | Palo Alto FW     |
        +------------------+
                |
            GWLBe
                |
        Transit Gateway
        /      |      \
     VPC-A   VPC-B   VPC-C
```