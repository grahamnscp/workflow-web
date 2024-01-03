db.createUser(
  {
    user: "bankuser",
    pwd: "bankuserpwd",
    roles: [
      {
      role: "readWrite",
      db: "bank"
      }
    ]
  }
);

db.createCollection('bankusers');
db.bankusers.createIndex({ username: 1 }, { unique: true }),
db.bankusers.createIndex({ email: 1 }),
db.bankusers.createIndex({ approvallevel: 1 }),
db.bankusers.insert(
  [
    { username: 'admin',     email: 'admin@examplebank.co',     approvallevel: 0 },
    { username: 'billsmith', email: 'billsmith@examplebank.co', approvallevel: 1 },
    { username: 'burt',      email: 'burt@examplebank.co',      approvallevel: 1 },
    { username: 'larrywell', email: 'larrywell@examplebank.co', approvallevel: 2 },
    { username: 'kayteller', email: 'kayteller@examplebank.co', approvallevel: 2 }
  ]
);

db.createCollection('creditblacklist');
db.creditblacklist.createIndex({ email: 1 }, { unique: true }),
db.creditblacklist.insert(
  [
    { email: 'sal@aol.us' },
    { email: 'nickgamble@bt.internet' },
    { email: 'halborrow@field.house' },
    { email: 'fellon@cellblack.h' },
    { email: 'failcredit@test' },
    { email: 'badcredit@test' }
  ]
);

db.createCollection('fraudrisk');
db.fraudrisk.createIndex({ email: 1 }, { unique: true }),
db.fraudrisk.insert(
  [
    { email: 'fred@home.telco',     risk: 2 },
    { email: 'casinoharry@lv.crew', risk: 7 },
    { email: 'fellon@cellblack.h',  risk: 10 },
    { email: 'fraud1@test',         risk: 1 },
    { email: 'fraud5@test',         risk: 5 },
    { email: 'fraud7@test',         risk: 7 },
    { email: 'fraud10@test',        risk: 10 }
  ]
);

