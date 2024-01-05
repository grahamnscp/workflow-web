# Standing Orders
Regular Payments from one account to another account, fixed amount with reference comment.  
  
![sorder-home-banner](../assets/sorder-home-banner.png)  
  
## Demo Architecture
![sorders-demo-architecture](../static/sorders-demo-architecture.png)  
![sorders-detailed-architecture](../static/sorders-detailed-architecture.png)
  
## Standing Orders UI
This demo just pays on a timer, amend is handled by temporal queries and signals to read and change current workflow variables.

### Create a new Standing Order
![sorder-new](../assets/sorder-new.png)  

### Create acknowledgement
![sorder-created](../assets/sorder-created.png)  

### List currently running standing order workflows
![sorder-list](../assets/sorder-list.png)  

### Standing Order details, Amend or Delete
![sorder-amend](../assets/sorder-amend.png)  

### Console output from Schedule worker
![sorder-worker](../assets/sorder-worker-console.png)  
  
