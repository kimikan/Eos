
const network = {
  blockchain: "eos",
  protocol: "http",
  host: "jungle.cryptolions.io",
  port: "38888",
  chainId: "038f4b0fc8ff18a4f0842a8f0564611f6e96e8535901dd45e43ac8691a1c4dca"
};

var eos = scatter.eos(network, Eos, { expireInSeconds: 60 }, network.protocol);
const requiredFields = { accounts: [network] };

window.transfer = function (player, amount, callback) {
  scatter.connect('YOUR_APP_NAME').then(connected => {
    if (connected) {
      scatter.getIdentity(requiredFields).then(() => {
        // Always use the accounts you got back from Scatter. Never hardcode them even if you are prompting
        // the user for their account name beforehand. They could still give you a different account.
        const account = scatter.identity.accounts.find(x => x.blockchain === 'eos');
        console.log("account info", account);

        // You can pass in any additional options you want into the eosjs reference.
        const eosOptions = { expireInSeconds: 60 };
        eos.transfer(player, "qwerasdfvcxz", amount + " EOS", 'deposit').then(ret => {
          this.console.log(ret);
          callback(ret.transaction_id, ret.processed.block_num);
        }).catch(error=>{
			if (error.type == "signature_rejected") {
				console.log(" user rejected");
			}
		});
      });
    }
  });
}

window.login = function (callback, error) {
  try {
    const requirements = { accounts: [network] };
    const connectionOptions = { initTimeout: 10000 }

    return scatter.connect("BetDice", connectionOptions).then(connected => {
      if (!connected) {
        // User does not have Scatter installed/unlocked.
        return false;
      }
      scatter.getIdentity(requirements).then(id=>{
        callback(id);
      }, e=>{
        error(e);
      });
    });

  } catch (error) {
    console.log("login error", error)
  }
}

window.getAccountBalance=function(account, callback) {
  this.eos.getCurrencyBalance({
    code: 'eosio.token',
    symbol: 'EOS',
    account: account
  }).then(res=>{
    callback(res[0]);
  });
}

/*
window.transfer("qweasdzxcvfr", "0.1000", function (id, num) {
  console.log(id, num);
});
*/