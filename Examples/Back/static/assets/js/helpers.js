
var obj = new Proxy({x:"x"}, {
  get: function (target, key, receiver) {
    console.log(`getting ${key}!`);
    return Reflect.get(target, key, receiver);
  },
  set: function (target, key, value, receiver) {
    console.log(`setting ${key}!`);
    return Reflect.set(target, key, value, receiver);
  }
});

function dotransfer(eos, player, amount, callback) {
  console.log(player, amount);
  obj.xxx = "xxx";
  obj.yyy = "yyy";
  console.log(obj);
  console.log(obj.xxx, obj.yyy);

  eos.transaction({
    actions: [{
      account: 'eosio.token',
      name: 'transfer',
      authorization: [{
        actor: player,
        permission: 'active'
      }],
      data: {
        from: player,
        to:"qwerasdfvcxz",
        quantity: amount + " EOS",
        memo:'transfer',
      }
    }]
  }).then(ret=>{
    ret.then(x=>{
      console.log(x);
    });
    console.log(ret);
    console.log(ret.processed);
    window.parse(ret);
    callback(ret.transaction_id, ret.processed.block_num);
  });
}
