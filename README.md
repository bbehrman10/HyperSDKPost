# Into the Avalanche: A Week with HyperSDK

## Introduction

Howdy folks, let's have a chat about virtual machines! 

As a developer passionate about blockchain technology, the power and potential of decentralized networks have always piqued by curiosity.  Because of that, I've spent a good deal of time writing applications & smart contracts that can run on these blockchains. But what about writing my own blockchain platform; my own digital playground upon which I can make the rules. Sounds daunting, right? Well, it doesn't necessarily have to be - and that's where Avalanche comes in. 

For those who don't know, Avalanche isn't like your typical blockchain platform - it's more like an ecosystem that is built up of multiple interoperable blockchain networks that communicate with each other. Developers can  spin up their own blockchain networks (called subnets) to run their dApps. This gives Avalanche a unique advantage against other Layer 1 blockchain networks when it comes to the speed and scaling of transactions. These subnets can be launched with a traditional EVM  or even a custom one. Now you're probably thinking - "Ben, creating custom virtual machines is a hard and complicated process." Okay, well maybe you weren't thinking that, but I can tell it was coming. It's true, there are many factors to consider when designing your own blockchain virtual machine. Much of that design process revolves around the blockchain network running properly and efficiently. That means you'd have to be a blockchain superstar to write your own machines, right?. Well what if I told you it didn't have to be that way? [Enter HyperSDK](https://github.com/ava-labs/hypersdk)

HyperSDK is Avalanche's own framework designed for building high-performance blockchain virtual machines. It can help developers navigate the *icy* landscape of blockchain development, taking care of the complex pieces and allowing them to focus on the unique aspects of their  applications.

In this post, I'll share my own experience with learning HyperSDK. After starting with its key features & components, I'll dive into a virtual machine I am working on, and hopefully help you gain an understanding of what HyperSDK is and why it might just be the snowmobile you need to ride the avalanche with your own virtual machines. Rev those engines, and let's get into it!

## Key Features

### State Management
When designing any virtual machine, state management is a critical element. With blockchains it's even more important, because you need to be able to efficiently track changes to balances, accounts, and other state objects within the chain. If proper state management isn't implemented, transaction performance and data management can greatly suffer. 

HyperSDK directly deals with these challenges by using a specially designed data structure - `x/merkledb` - a path-based merkelized radix tree implementation. This high-performance data structure minimizes the on-disk footprint of any HyperVM by deleting any data that is no longer part of the current state, without performing any costly reference counting.

This approach gives HyperVMs a whole new level of scalability and design potential. The merkledb both reduces the amount of disk space required and improves transaction processing times. This let's us as developers focus more on the unique aspects of our applications rather than worrying about performance and scalability issues. 

Another state management feature of HyperSDK is dynamic state sync. Normally, when a new node joins a blockchain network, it has to download, execute, and then verify all previous transactions on chain in order to catch up. In a HyperChain, a blockchain built using a HyperVM, when a new node joins, it leverages a package called `x/sync`. This allows nodes to download the latest state of the network, thereby significantly reducing the time and resources needed for synchronization. Furthermore, `x/sync` prevents nodes from falling behind during this process. This combination of efficient bandwidth usage and dynamic synchronization greatly enhances the performance and scalability of HyperChain networks and keeps them *cool* under pressure.

The final piece of the state management puzzle is `PebbleDB.` Many Avalanche VM implementations use `goleveldb` for storage which ends up impacting the total throughput, but HyperSDK instead uses CockroachDBs Pebble database for on disk storage. Rather than keeping data in the AvalancheGo root directory, Pebble stores the data under a set of distinct paths in the AvalancheGo chainData directory. This structure gives us the ability to run multiple disk drives when running a HyperVM so we get access to more throughput. 

For more information on these two, visit: <br />
[x/merkledb](https://github.com/ava-labs/avalanchego/blob/master/x/merkledb/README.md)   <br />
[PebbleDB](https://github.com/cockroachdb/pebble#advantages)

### Block Execution

Block execution refers to the process of validating and adding new blocks of transactions to a blockchain. This is another critical operation in any blockchain system and often forms a bottleneck that limits the scalability and performance of the network. The HyperSDK framework was built to give developers access to "hyper" speed and scalability. It offers features, like state pre-fetching and parallel signature verification, that help optimize block execution.

 1. State Pre-Fetching
	  - All transactions with a HyperVM must contain the keys to the state they will either read or overwrite during both execution and authentication. This gives the virtual machine the ability to pre-fetch relevant data for a transaction before block execution even starts. This gives that process access to all the verification data it needs via memory. This will eventually give HyperVMs the ability to process transactions in a parallel manner (this feature is disabled at the time of writing).
 
 2. Parallel Signature Verification
	 - HyperSDK supports parallel signature verification. This means that the verification of signatures, which is a computationally intensive process, can be performed concurrently for different transactions. This reduces the end-to-end verification time of a block when running on powerful hardware, thus speeding up block execution.

### Account Abstraction

Account abstraction is another notable feature in the HyperSDK framework. The concept of account abstraction essentially refers to the framework making no assumptions about how transactions (also known as `Actions`) are verified on a HyperChain.

Instead of imposing a specific validation method, HyperSDK provides a registry of supported authentication modules that can be used to validate each type of transaction. These `Auth` modules can perform a variety of tasks, ranging from simple operations like signature verification to more complex tasks like executing a WebAssembly (WASM) blob.

This design choice offers a lot of flexibility and customizability to developers, allowing us to define the verification process that best suits the needs for each type of transaction on a HyperChain. This both enhances the security of the blockchain and potentially opens up new possibilities for innovative transaction types and interactions on the blockchain​.

### Nonceless Transactions

In the world of account-based blockchains, there's a need to protect against replay attacks, which are a type of network attack where an attacker tries to repeat or delay a valid transaction. Many blockchains use nonces (a number that can only be used once) to protect agains these types of attacks. A nonce is required to be incremented each time a transaction occurs so that each transaction is sent and processed in the correct order. This, however, can be inconvenient for users because if they do not account for the nonce, a transaction can be delayed or even dropped; holding up all subsequent transactions for their account.

HyperSDK takes a different approach to transaction security. Transactions contain a time parameter which essentially limits how long a transaction can be included inside of a block (an expiration date for transactions). This removes the need to use nonces. And without nonces, users can do a lot more things with their transactions such as submitting multiple transactions concurrently. It's a feature that not only simplifies the user process, but also provides significant performance benefits. Without nonces, the mempool (the set of unconfirmed transactions) will be more performant as there isn't a need to maintain multiple transactions for a single account or to make sure they are ordered. It provides benefit to the network layer as any valid transaction can be propagated to any node instead of just the ones that can be executed at that exact moment in time. It also means: no more replacement transactions! If a user wants to cancel a transaction, they can just wait for the expiration date on the transaction to expire.

Overall, the nonce-less transaction model of HyperSDK offers both usability and performance benefits, making it a compelling choice for building scalable, highly-performative blockchains.

### Warp Messaging

Every HyperVM has the ability to use Avalanche Warp Messaging (AWM) out-of-the-box. AWM is a primitive tool provided by the Avalanche Network that allows any subnet to send and receive messages to any other subnet on the Avalanche Network. This feature is crucial to subnet communication and allows for a truly inter-connected network of subnets. Read more about AWM [here](https://support.avax.network/en/articles/6846276-what-is-avalanche-warp-messaging-awm).

### Easy Functionality Upgrades
Every single on-chain object and parameter is scoped by block timestamp. That means the rules, properties, and even functionality for any of these can be changed based on the timestamp of the block they are in. Need to make an upgrade to your VM? No sweat! Need to disable a feature for a specific amount if time? Easy! Your blockchain just became fully upgradable.

### Proposer Aware Gossip
Traditional Avalanche virtual machine gossip can often feel like that overly enthusiastic friend who will broadcast the smallest piece of news to everyone they know. While, yes, this ensures the information gets around it can also lead to a ton of unnecessary noise. That's where proposer aware gossip comes in similarly to a trusted confidant to whom we can share information with knowing it'll be disclosed judiciously. Since we prioritize the proposer (the key decision maker for the next block), it insures the right information will reach the right place while also reducing overall gossip further improving network efficiency. 

### Transaction Results & Execution Rollback
For a moment, imagine you're putting together a jigsaw puzzle and you're testing a piece out to see if it fits. You put the piece in and it doesn't. The piece is now stuck there forever and you're unable to remove it. That is what it is like to work with a blockchain  without transaction results or execution rollback. HyperSDK allows for actions in a blockchain to have outcomes (results, fees, and success statuses). If an action is unsuccessful, any changes it made to the blockchain's state are reverted. This feature is useful for developing smart contract-based systems that might need to stop execution early if certain conditions aren't met, which can be more cost-effective than carrying out the full transaction.

However, HyperSDK does not record or index the results of these actions. Unlike many other blockchain tools, it doesn't have an "archival mode" for historical access. Instead, it only keeps the information needed to validate the next block in the chain and to help new nodes synchronize with the current state of the blockchain. When a block is accepted, a HyperVM will communicate the results of all executions upon which we could perform a callback function if we want that information saved for later.

### Unified Metrics & Logging
Unified logging and metrics are also included within HyperSDK. Its logging works by collecting log data from various parts of network and consolidating it into a single, manageable stream. This log data can include everything from transaction processing times to system errors, providing a comprehensive view of a HyperChain's activities. The unified nature of this logging system makes it easier to identify patterns or issues that might otherwise be missed when data is spread across multiple, disparate logs.

Phew! That was a lot of information. Let's take a moment to digest this. It's clear there's a wealth of functionality and advanced optimization techniques at work beneath the surface of HyperSDK. While intricate and fascinating, they are largely managed by HyperSDK itself freeing us from the burden of having to do it. So now we can shift our focus towards the more practical and tangible components of a HyperVM. 

##  Components of a HyperVM

In this section we'll go through each interface of a HyperVM. 

### Controller 
The `Controller` is a main component of a HyperVM. As the entry point, it is responsible for initializing all the data structures that will be utilized. These data structures include configuration files, builders, gossipers, databases, handlers, and registries.

This interface has several methods:

-   `Initialize`: This method sets up the virtual machine with a variety of parameters including the VM itself, context, metric gatherer, genesis bytes, upgrade bytes, and configuration bytes. It returns a set of key components for the VM, including the config, genesis, builder, gossiper, vmDB, stateDB, handler, actionRegistry, authRegistry, or an error if one occurs.
-   `Rules`: This method retrieves the chain rules.
-   `StateManager`: This method retrieves the state manager.
-   `Accepted` and `Rejected`: These methods handle Accepted and Rejected block callbacks, respectively.
-   `Shutdown`: This method manages the process of shutting down the context.

Most HyperVMs use the default `Builder`, `Gossiper`, `Handlers`, and `Database` packages, so the `Controller` code is often boilerplate.

The `Controller` also provides `ActionRegistry` and `AuthRegistry`. These registries inform the HyperSDK how to marshal (convert into a byte stream) and unmarshal (convert back into the original data) bytes on-the-wire. In other words, they provide instructions for HyperSDK to interpret the bytes it receives from the Avalanche Consensus Engine. If these registries were not provided, HyperSDK would not be able to extract any information from the byte data it receives.

Take a look at the [controller for Avalanche's `tokenvm`](https://github.com/ava-labs/hypersdk/blob/main/examples/tokenvm/controller/controller.go)  to see it in action. 

### Genesis 
The `Genesis` interface in HyperSDK represents the initial state of a blockchain when it is first instantiated. It includes elements like the starting balances of accounts and various default configurations, such as the price of transaction fees or the types of transactions that are initially enabled.

In practice, the `Genesis` interface will be implemented differently depending on the specific requirements of the blockchain being created. For example, if you were creating a blockchain for a token economy, the `Genesis` might include the initial distribution of tokens to specific addresses. If you were creating a blockchain for a decentralized voting system, the `Genesis` could include the initial list of eligible voters or candidates. [Here is Avalanche's `tokenvm` implementation of the `Genesis` interface](https://github.com/ava-labs/hypersdk/blob/main/examples/tokenvm/genesis/genesis.go)

Regardless of the specifics, the genesis state is always stored on the Avalanche Platform Chain (P-Chain) when the blockchain network is created. This allows anyone to verify the initial state of the blockchain.

### Actions
An `Action` in the context of the HyperSDK is an interface that defines how users interact with the blockchain runtime. It's essentially the user-defined element of any HyperSDK transaction that gets processed by all participants of a HyperChain. It  encapsulates the rules and operations that govern the behavior of transactions in the HyperSDK environment.

The `Action` interface has several key required methods:
-   `MaxUnits`: This method is used to specify the maximum units that an action can consume.
-   `ValidRange`: This method defines the valid temporal range for the action.
-   `StateKeys`: This method provides the state keys that the action will touch during execution.
-   `Execute`: This method is where the actual execution of the action takes place. It involves the application of the rules, interaction with the database, and processing of the transaction.
-   `Marshal`: This method is used for serialization of the action.
-    `Unmarshal`: This method is used for unserializing an action

The output of an `Action`'s `Execute` method is a `Result`, which indicates whether the execution was successful, how many units were used, an output (which could be arbitrary bytes specific to the HyperVM), and potentially a `WarpMessage` which Subnet Validators will sign.

Examples of `Action` could include transferring tokens from one account to another, creating a smart contract, updating the state of a smart contract, or any other operation that you might typically perform in a blockchain environment. The specific implementation of these methods would vary greatly depending on the particular use case and the rules of the specific blockchain being developed.

For instance, in a simple token transfer `Action`, the `Execute` method might involve checking the sender's balance, subtracting the transfer amount from the sender's account, and adding it to the recipient's account. The `StateKeys` method would return the keys related to the sender's and receiver's account balances. The `MaxUnits` method could be used to limit the amount of tokens that can be transferred in a single transaction, and the `ValidRange` method could be used to specify a time period during which the transfer can occur.

### Auth
The `Auth` interface allows developers to create custom authentication rules that align with the specific needs or goals of their blockchain project. It has a few similarities with the `Action` interface. They both are utilized by HyperSDK to define specific actions or behaviors within the blockchain. However, the `Auth` interface adds the concept of a "payer" who is responsible for the fees associated with the operations taking place in an `Action`. Any fees that aren't used can be returned to the payer if specified in the corresponding `Action` that was authenticated.

For instance, an `Auth` could be set up to allow users to rotate their keys or enable others to perform specific actions on their behalf. It could also be structured to let accounts natively pay for the fees of other accounts. These features can be particularly beneficial for server-based accounts that want to implement a periodic key rotation scheme without losing their on-chain history, which can affect their reputation.

### Rules
The `Rules` interface in HyperSDK governs the validity of blocks within a blockchain. It defines various parameters related to the blockchain's operation, such as the maximum number of transactions per block, the maximum block size, the block's validity window, and the base units of the blockchain. It also includes rules for the unit price, block cost, and warp configuration.

These rules are requested from the Controller before executing any block, and the HyperVM performs this request so the Controller can modify any rules on-the-fly. This dynamic adjustment of rules can be essential for maintaining optimal performance and security of the blockchain network.

Moreover, the `Rules` interface provides an option to define custom rules that can be accessed during Auth or Action execution. This feature can be especially useful for implementing project-specific rules or behaviors.

### Storage 
The `Storage` interface for HyperSDK provides an abstraction layer for managing and querying various types of data related to a blockchain-based application. The module interacts with a key-value database to store and retrieve data, and it organizes the data using key prefixes and unique identifiers. The module may be used by other components of the HyperSDK-based application to access and manipulate data in a structured manner and can even be used to disseminate Avalanche Warp Messages.

A typical storage module for HyperSDK might include:

1.  Key Prefixes & Construction: To facilitate organized storage and retrieval of data, it's a good idea to define a set of constant key prefixes to categorize different types of data. Then a constructor function can use those prefixes alongside elements of a given transaction to generate a unique key.
    
2.  Data Management: Depending on what types of actions you are doing, it will be important to provide functions for managing different types of data such as transactions, account balances, orders, and any other application specific data. These functions allow CRUD commands on these data types.
    
3.  Error Handling: It also defines any errors that can occur during database operations that can be provided back to the caller in the event of an error.    

Now that we've covered the fundamental components of a HyperVM and how they work together to facilitate high-performance blockchains on Avalanche, let's pivot to a practical use case. This will give you an idea of what some of the code for the modules I mentioned above would look like for a specific use case.

## Use Case:  Energy Trading
Energy grid management and facilitation has a great deal of potential for innovation when it comes to decentralized networks. In this use case, I've built out a few components of a HyperVM that will facilitate a simplified energy grid use case. I have found it best to start with what actions you want your chain to execute and build the other components out from there.

### Actions
`initialize_energy_asset` - this will create the energy asset, it will only need to be called at the chain's initialization as there isn't a need for more than 1 asset type: Energy - represented generally in kilowatt hours.

`produce_energy`- this will be called by participants that generate energy - think power companies, but also individuals who 	generate their own electricity be that solar panels or anything else. Very similar to minting an asset. 

`consume_energy` - this will be called by participants when they consume energy it will burn the energy asset to keep the chain's supply in line with the grid's.

`create_energy_order` - this will be called by participants who wish to sell their excess energy. They'll set aside an amount of energy at a specified price point that other participants can then buy energy. 

`fill_energy_order` - this is called by participants who want to buy energy. They'll pick an order that was set in the previous action and swap out the native chain's token for the energy asset.

`close_energy_order` - producers will call this function to close out a previously created order

Now, each of these actions are their own file within the `actions` folder within the HyperVM project. It's important to remember including the key methods for each action.  You'll see that in action below in my `initialize_energy_asset.go` file:

    var  _ chain.Action = (*InitializeEnergyAsset)(nil)  
    type  InitializeEnergyAsset  struct {
	    Metadata []byte  `json:"metadata"`
	}  
    func (*InitializeEnergyAsset) StateKeys(_ chain.Auth, txID ids.ID) [][]byte {
	    return [][]byte{storage.PrefixAssetKey(txID)}
    }  
    func (c *InitializeEnergyAsset) Execute(
	    ctx context.Context,
	    r chain.Rules,
	    db chain.Database,
	    _ int64,
	    rauth chain.Auth,
	    txID ids.ID,
	    _ bool,
    ) (*chain.Result, error) {
	    actor  := auth.GetActor(rauth)
	    unitsUsed  := c.MaxUnits(r)
	    if  len(c.Metadata) > MaxMetadataSize {
		    return  &chain.Result{Success: false, Units: unitsUsed, Output: OutputMetadataTooLarge}, nil
	    }
	    if  err  := storage.SetAsset(ctx, db, txID, c.Metadata, 0, actor, false); err !=  nil {
		    return  &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	    }
	    return  &chain.Result{Success: true, Units: unitsUsed}, nil
    }  
    func (c *InitializeEnergyAsset) MaxUnits(chain.Rules) uint64 {
	    return  uint64(len(c.Metadata))
    }  
    func (c *InitializeEnergyAsset) Marshal(p *codec.Packer) {
	    p.PackBytes(c.Metadata)
    }  
    func  UnmarshalCreateAsset(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	    var  create InitializeEnergyAsset
	    p.UnpackBytes(MaxMetadataSize, false, &create.Metadata)
	    return  &create, p.Err()
    }
	func (*InitializeEnergyAsset) ValidRange(chain.Rules) (int64, int64) {
	    return  -1, -1
    }

### Authorization
To keep the authorization simple. I'm going to use the same standard ED25519 authorization that [Avalanche provides in their `tokenvm`](https://github.com/ava-labs/hypersdk/blob/main/examples/tokenvm/auth/ed25519.go) and make a couple adjustments to account for the use case. We're going to include some attributes in the struct to account for tracking energy produced and consumed. 

    type  ED25519  struct {
	    Signer crypto.PublicKey `json:"signer"`
	    Signature crypto.Signature `json:"signature"`
	    Consumption float64  `json:"consumption"`
	     Production float64  `json:"production"`
	    TimestampUTC time.Time `json:"timestamp_utc"`
     }

I will also wrote some set functions that are responsible for setting the values of these attributes.

    func (d *ED25519) SetConsumption(consumption float64) {
	    d.Consumption  = consumption
    }  
    func (d *ED25519) SetProduction(production float64) {
	    d.Production  = production
    }  
    func (d *ED25519) SetTimestampUTC(timestamp time.Time) {
	    d.TimestampUTC  = timestamp
    }  
    func (d *ED25519) AdjustConsumption(consumptionDelta float64) {
	    d.Consumption += consumptionDelta
    }  
    func (d *ED25519) AdjustProduction(productionDelta float64) {
	    d.Production += productionDelta
    }  
    func (d *ED25519) GetExcessProduction() float64 {
	    if d.Production > d.Consumption {
		    return d.Production - d.Consumption
	    }
    return  0
    }

With that we can move onto the registry.

### Registry
From here we can go into our registry directory and make a `registry.go` file. The main piece of the registry file is to "register" all the different actions and authentication methods used on your chain. We start my defining our two registries. Note we are using HyperSDK's provided codec & chain libraries here.
    
    consts.ActionRegistry = codec.NewTypeParser[chain.Action, *warp.Message]()    
    consts.AuthRegistry = codec.NewTypeParser[chain.Auth, *warp.Message]()
        
Next we define our actions. using the built in Register functions. 
   
    consts.ActionRegistry.Register(&actions.CreateEnergyAsset{}, actions.UnmarshalCreateAsset, false),    
    consts.ActionRegistry.Register(&actions.ProduceEnergy{}, actions.UnmarshalProduceEnergy, false),    
    consts.ActionRegistry.Register(&actions.ConsumeEnergy{}, actions.UnmarshalConsumeEnergy, false),    
    consts.ActionRegistry.Register(&actions.CreateEnergyOrder{}, actions.UnmarshalCreateEnergyOrder, false),    
    consts.ActionRegistry.Register(&actions.FillEnergyOrder{}, actions.UnmarshalFillEnergyOrder, false),    
    consts.ActionRegistry.Register(&actions.CloseEnergyOrder{}, actions.UnmarshalCloseEnergyOrder, false),              
    
Next we register any authorization modules

     consts.AuthRegistry.Register(&auth.ED25519{}, auth.UnmarshalED25519, false),
        
### Genesis
With the registry complete, it's time to make our `genesis.go` file within the genesis directory. I used one very similar to the [Avalanche provided `tokenvm` virtual machine](https://github.com/ava-labs/hypersdk/blob/main/examples/tokenvm/genesis/genesis.go) renaming `balance` to `energy` within the custom allocation. Everything else is the same.

    type  CustomAllocation  struct {
    Address string  `json:"address"`  // bech32 address
    Energy uint64  `json:"kilowattHours"`
    }

This directory is also where we set any rules for our chain. [I'm using these boilerplate rules](https://github.com/ava-labs/hypersdk/blob/main/examples/tokenvm/genesis/rules.go).

### Storage
Now let's define the functions that will facilitate the storage of our data. First we define our transaction prefixes:

    txPrefix = 0x0
	balancePrefix = 0x1
	assetPrefix = 0x2
	energyOrderPrefix = 0x3
	creditPrefix = 0x4
	heightPrefix = 0x5
	incomingWarpPrefix = 0x6
	outgoingWarpPrefix = 0x7
	
Each type of transaction will warrant it's own helper function to be called by the storage methods. Here is what one of those looks like:

    func  PrefixEnergyOrderKey(txID ids.ID) (k []byte) {
	    k = make([]byte, 1+consts.IDLen)
	    k[0] = energyOrderPrefix
	    copy(k[1:], txID[:])
	    return
    }

In this snippet we are taking our passed in transaction ID and appending that onto our energyOrderPrefix constant and returning it. Any transaction that has to do with Energy Orders will call this function. 

Let's take a look at the function used to write an Energy Order into the database:

    func  SetEnergyOrder(
    	ctx context.Context,
    	db chain.Database,
    	tdID ids.ID,
    	in ids.ID,
    	inTick uint64,
    	out ids.ID,
    	outTick uint64,
    	supply uint64,
    	owner crypto.PublicKey,
    ) error {
    	k := PrefixEnergyOrderKey(tdID)
    	v := make([]byte, consts.IDLen*2+consts.Uint64Len*3+crypto.PublicKeyLen)
    	copy(v, in[:])
    	binary.BigEndian.PutUint64(v[consts.IDLen:], inTick)
    	copy(v[consts.IDLen+consts.Uint64Len:], out[:])
    	binary.BigEndian.PutUint64(v[consts.IDLen*2+consts.Uint64Len:], outTick)
    	binary.BigEndian.PutUint64(v[consts.IDLen*2+consts.Uint64Len*2:], supply)
    	copy(v[consts.IDLen*2+consts.Uint64Len*3:], owner[:])
    	return db.Insert(ctx, k, v)
    }

The `SetEnergyOrder` function takes a number of different parameters:
`ctx` - we pass in context as a matter of best practices
`db`- this is the database we are going to use. In this function we use the built in HyperSDK chain.Database (PebbleDB)
`txID` - an identifier for the order
`in`- represents the asset that is paying the transaction
`inTick`- the amount of the `in` asset used
`out`- represents the asset to be filled with this order, in our case it will be the energy asset
`outTick`- the amount of the energy asset to be traded for given the `in` asset
`supply`- the total amount of energy that can be filled from this order
`owner` - the owner or producer of energy

The function starts by calling the PrefixEnergyOrderKey function to construct a database key. It then creates a byte slice with enough space to hold the information for the order. The length is calculated based on the sizes of the various attributes (e.g., `consts.IDLen` for asset IDs, `consts.Uint64Len` for 64-bit integers, `crypto.PublicKeyLen` for public keys).

Then using the `encoding/binary` file, we serialize our transaction data and eventually call `db.Insert` passing in the context and the key and value pair. 

Heres a full list of all the functions I have in this file:      <br />
`PrefixTxKey()`: Generates a database key for a transaction using the transaction ID. <br />
`PrefixBalanceKey()`: Generates a database key for an account balance using the public key and asset ID.     <br />
`PrefixAssetKey()`: Generates a database key for an asset using the asset ID.      <br />
`PrefixEnergyOrderKey()`: Generates a database key for an energy order using the order ID.     <br />
`CreditPrefixKey()`: Generates a database key for a credit entry using the asset ID and destination ID.     <br />
`HeightKey()`: Generates a database key for the blockchain height.     <br />
`IncomingWarpKeyPrefix(sourceChainID ids.ID, msgID ids.ID) []byte`: Generates a database key for an incoming warp message using the source chain ID and message ID.     <br />
`OutgoingWarpPrefix()`: Generates a database key for an outgoing warp message using the transaction id <br />
`StoreTransaction()`: Stores a transaction in the database, including its ID, timestamp, success status, and units. <br />
`GetTransaction()`: Retrieves a transaction from the database using its ID. <br />
`GetBalance()`: Retrieves the balance of an account for a specific asset. <br />
`SetBalance()`: Sets the balance of an account for a specific asset. <br />
`DeleteBalance()`: Deletes the balance entry of an account for a specific asset. <br />
`AddBalance()`: Adds an amount to the balance of an account for a specific asset. <br />
`SubBalance()`: Subtracts an amount from the balance of an account for a specific asset. <br />
`GetAssetFromState()`: Retrieves asset information from the state using a ReadState function. <br />
`GetAsset()`: Retrieves asset information from the database using the asset ID. <br />
`SetAsset()`: Stores asset information in the database, including metadata, supply, owner, and warp status. <br />
`SetEnergyOrder()`: Stores an energy order in the database, including its attributes such as input asset, output asset, supply, and owner. <br />
`GetEnergyOrder()`: Retrieves an energy order from the database using the order ID. <br />
`DeleteOrder()`: Deletes an energy order entry from the database using the order ID. <br />
`GetCreditFromState()`: Retrieves credit information from the state using a ReadState function. <br />
`GetCredit()`: Retrieves credit information from the database using the asset ID and destination ID. <br />
`SetCredit()`: Stores credit information in the database, including the amount of credit. <br />
`AddCredit()`: Adds an amount to the credit entry in the database. <br />
`SubCredit()`: Subtracts an amount from the credit entry in the database. <br />

### Controller
There isn't a huge amount we need to change here from the boilerplate. Mostly we need to adjust the `Accepted` function's success path to include some metrics we define in an adjacent file.

	if result.Success {
		switch  action := tx.Action.(type) {
		case *actions.InitializeEnergyAsset:
			c.metrics.InitializeEnergyAsset.Inc()
		case *actions.ProduceEnergy:
			c.metrics.ProduceEnergy.Inc()
		case *actions.ConsumeEnergy:
			c.metrics.ConsumeEnergy.Inc()
		case *actions.CreateEnergyOrder:
			c.metrics.createEnergyOrder.Inc()
		actor := auth.GetActor(tx.Auth)
			c.energyLedger.Add(tx.ID(), actor, action)
		case *actions.FillEnergyOrder:
			c.metrics.fillEnergyOrder.Inc()
			energyOrderResult, err := actions.UnmarshalOrderResult(result.Output)
			if err != nil {
				return err
			}
			if orderResult.Remaining == 0 {
				c.orderBook.Remove(action.Order)
				continue
			}
			c.energyLedger.UpdateRemaining(action.Order, orderResult.Remaining)
		case *actions.CloseEnergyOrder:
			c.metrics.closeEnergyOrder.Inc()
			c.energyLedger.Remove(action.Order)
	}}

Then the `metrics` file:

    type  metrics  struct {
    	initializeEnergyAsset prometheus.Counter
    	produceEnergy prometheus.Counter
    	consumeEnergy prometheus.Counter
    	createEnergyOrder prometheus.Counter
    	fillEnergyOrder prometheus.Counter
    	closeEnergyOrder prometheus.Counter 
    }  
    func  newMetrics(gatherer ametrics.MultiGatherer) (*metrics, error) {
    	m := &metrics{
    		initializeEnergyAsset: prometheus.NewCounter(prometheus.CounterOpts{
    		Namespace: "actions",
    		Name: "initialize_energy_asset",
    		Help: "number of initialize asset actions",
    	}),
    		produceEnergy: prometheus.NewCounter(prometheus.CounterOpts{
    		Namespace: "actions",
    		Name: "produceEnergy",
    		Help: "number of produce energy actions",
    	}),
    		consumeEnergy: prometheus.NewCounter(prometheus.CounterOpts{
    		Namespace: "actions",
    		Name: "consumeEnergy",
    		Help: "number of consume energy actions",
    	}),
    		createEnergyOrder: prometheus.NewCounter(prometheus.CounterOpts{
    		Namespace: "actions",
    		Name: "create_energy_order",
    		Help: "number of create order actions",
    }),
    		fillEnergyOrder: prometheus.NewCounter(prometheus.CounterOpts{
    		Namespace: "actions",
    		Name: "fill_energy_order",
    		Help: "number of fill energy order actions",
    }),
    		closeEnergyOrder: prometheus.NewCounter(prometheus.CounterOpts{
    		Namespace: "actions",
    		Name: "close_energy_order",
    		Help: "number of close energy order actions",
    }),
    }
    r := prometheus.NewRegistry()
    errs := wrappers.Errs{}
    errs.Add(
    r.Register(m.initializeEnergyAsset),
    r.Register(m.produceEnergy),
    r.Register(m.consumeEnergy), 
    r.Register(m.createEnergyOrder),
    r.Register(m.fillEnergyOrder),
    r.Register(m.closeEnergyOrder),  
    gatherer.Register(consts.Name, r),
    )
    return m, errs.Err
    }
    
In this file we utilize the HyperSDK metrics module to help us register the metrics for each action which are then counted using Prometheus.

### Energy Ledger
Similar to the `orderbook` inside of `tokenvm`, I've created a similar module - `energyledger` to keep track of all the different energy orders that are made. You can [view my `energyledger` code here.](https://github.com/bbehrman10/HyperSDKPost/blob/main/Code/energyledger/energyledger.go)

Those are most of the basic components we will need for this HyperVM. [You can view the full set of code I put together here.](https://github.com/bbehrman10/HyperSDKPost/tree/main/Code) 

## Closing Thoughts & Next Steps

In conclusion, HyperSDK offers a revolutionary way to build scalable blockchains on Avalanche. It offers an abstraction layer that reduces the complexity of building your own blockchain runtime, allowing developers to focus on the unique aspects of their project. HyperSDK's opinionated design methodology means that most runtimes built on it only need to implement a minimal amount of their own code to add custom interaction patterns, saving developers significant time and effort. After just a little over a week with this framework I can firmly say that the possibilities are truly endless. 

For my own HyperVM journey, I have these next steps mapped out:

 1. Flesh out this VM skeleton
 2. Launch the virtually machine locally via the `avalanche-cli` 
 3. Expand the feature set of the EnergyVM
 
As of now this code will not quite yet run. For starters, I need to include an RPC server module. If you look through my code you'll see references to it, but I have yet to finish that module. Once all the components are finalized, I can move onto launching this on a local network. For that, I will be using the `avalanche-cli`.  Then, it will be a great time to add more features to the virtual machine. I'd like to add interfaces for IoT devices, such as energy meters or batteries, to help automate the consuming and producing energy actions on chain. You wouldn't want to have to, for instance, manually transact each time you turn on a light switch. Another feature I could see being valuable is differentiating between the type of energy. 

I'll go through all of this and more in my next post. In the meantime I invite you to [watch this recording](https://www.youtube.com/watch?v=dAejw4AWK8I) of Patrick O'Grady and Dr Emir Gün Sirer discussing HyperSDK. I found it quite valuable when getting started with this framework.

Until next time,       
Ben B
