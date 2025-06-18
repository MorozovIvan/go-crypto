<template>
  <div class="p-8">
    <h2 class="mb-4 text-2xl font-bold">Find Profitable Solana Wallets</h2>
    <!-- Filters Bar -->
    <div class="flex flex-wrap items-end gap-4 mb-4">
      <div>
        <label class="block mb-1 text-xs font-semibold">Winrate</label>
        <div class="flex items-center gap-2">
          <input type="range" min="0" max="100" v-model.number="filters.winrate" class="w-32" />
          <span class="text-sm">{{ filters.winrate }}%</span>
        </div>
      </div>
      <div>
        <label class="block mb-1 text-xs font-semibold">Profit (Min)</label>
        <input type="number" v-model.number="filters.profitMin" class="w-24 px-2 py-1 border rounded" placeholder="Min" />
      </div>
      <div>
        <label class="block mb-1 text-xs font-semibold">Profit (Max)</label>
        <input type="number" v-model.number="filters.profitMax" class="w-24 px-2 py-1 border rounded" placeholder="Max" />
      </div>
      <div>
        <label class="block mb-1 text-xs font-semibold">#Tx (Min)</label>
        <input type="number" v-model.number="filters.txMin" class="w-20 px-2 py-1 border rounded" placeholder="Min" />
      </div>
      <div>
        <label class="block mb-1 text-xs font-semibold">#Tx (Max)</label>
        <input type="number" v-model.number="filters.txMax" class="w-20 px-2 py-1 border rounded" placeholder="Max" />
      </div>
      <div>
        <label class="block mb-1 text-xs font-semibold">Last Active (from)</label>
        <input type="date" v-model="filters.lastActiveFrom" class="px-2 py-1 border rounded" />
      </div>
      <div>
        <label class="block mb-1 text-xs font-semibold">Last Active (to)</label>
        <input type="date" v-model="filters.lastActiveTo" class="px-2 py-1 border rounded" />
      </div>
      <div>
        <label class="block mb-1 text-xs font-semibold">Token</label>
        <select v-model="filters.token" class="px-2 py-1 border rounded">
          <option value="">Any</option>
          <option v-for="token in allTokens" :key="token" :value="token">{{ token }}</option>
        </select>
      </div>
    </div>
    <div class="mb-4">
      <input
        v-model="search"
        type="text"
        placeholder="Search by wallet address..."
        class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#0a78b9]"
      />
    </div>
    <div class="p-6 bg-white rounded-lg shadow">
      <table class="min-w-full divide-y divide-gray-200">
        <thead>
          <tr>
            <th class="px-4 py-2 text-xs font-medium text-left text-gray-500 uppercase cursor-pointer" @click="sortBy('rank')">
              Rank
              <span v-if="sortKey === 'rank'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="px-4 py-2 text-xs font-medium text-left text-gray-500 uppercase cursor-pointer" @click="sortBy('address')">
              Wallet Address
              <span v-if="sortKey === 'address'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="px-4 py-2 text-xs font-medium text-left text-gray-500 uppercase cursor-pointer" @click="sortBy('profit')">
              Profit (SOL)
              <span v-if="sortKey === 'profit'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="px-4 py-2 text-xs font-medium text-left text-gray-500 uppercase cursor-pointer" @click="sortBy('pnl')">
              PNL
              <span v-if="sortKey === 'pnl'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="px-4 py-2 text-xs font-medium text-left text-gray-500 uppercase cursor-pointer" @click="sortBy('txCount')">
              #Tx
              <span v-if="sortKey === 'txCount'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="px-4 py-2 text-xs font-medium text-left text-gray-500 uppercase cursor-pointer" @click="sortBy('avgTokensPerDay')">
              Avg Tokens/Day
              <span v-if="sortKey === 'avgTokensPerDay'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="px-4 py-2 text-xs font-medium text-left text-gray-500 uppercase cursor-pointer" @click="sortBy('avgHoldTime')">
              Avg Hold Time (days)
              <span v-if="sortKey === 'avgHoldTime'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="px-4 py-2 text-xs font-medium text-left text-gray-500 uppercase cursor-pointer" @click="sortBy('lastActive')">
              Last Active
              <span v-if="sortKey === 'lastActive'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="px-4 py-2 text-xs font-medium text-left text-gray-500 uppercase cursor-pointer" @click="sortBy('winrate')">
              Winrate
              <span v-if="sortKey === 'winrate'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="px-4 py-2 text-xs font-medium text-left text-gray-500 uppercase cursor-pointer" @click="sortBy('tokens')">
              Token Holdings
              <span v-if="sortKey === 'tokens'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="px-4 py-2"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(wallet, i) in sortedWallets" :key="wallet.address" class="hover:bg-gray-50">
            <td class="px-4 py-2">{{ i + 1 }}</td>
            <td class="px-4 py-2 font-mono">{{ wallet.address }}</td>
            <td class="px-4 py-2 font-semibold text-green-700">{{ wallet.profit }}</td>
            <td class="px-4 py-2" :class="wallet.pnl >= 0 ? 'text-green-600' : 'text-red-600'">{{ wallet.pnl > 0 ? '+' : '' }}{{ wallet.pnl }}</td>
            <td class="px-4 py-2">{{ wallet.txCount }}</td>
            <td class="px-4 py-2">{{ wallet.avgTokensPerDay }}</td>
            <td class="px-4 py-2">{{ wallet.avgHoldTime }}</td>
            <td class="px-4 py-2">{{ wallet.lastActive }}</td>
            <td class="px-4 py-2">{{ wallet.winrate !== undefined ? wallet.winrate + '%' : '-' }}</td>
            <td class="px-4 py-2 font-mono">{{ wallet.tokens.join(', ') }}</td>
            <td class="px-4 py-2">
              <button @click="selectWallet(wallet)" class="text-[#0a78b9] hover:underline">View</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="sortedWallets.length === 0" class="py-8 text-center text-gray-400">No wallets found.</div>
    </div>

    <!-- Wallet Details Modal -->
    <div v-if="selectedWallet" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-40">
      <div class="relative w-full max-w-lg p-8 bg-white rounded-lg shadow-lg">
        <button @click="selectedWallet = null" class="absolute text-gray-400 top-2 right-2 hover:text-gray-700">&times;</button>
        <h3 class="mb-2 text-xl font-bold">Wallet Details</h3>
        <div class="mb-2"><span class="font-semibold">Address:</span> <span class="font-mono">{{ selectedWallet.address }}</span></div>
        <div class="mb-2"><span class="font-semibold">Profit:</span> <span class="font-semibold text-green-700">{{ selectedWallet.profit }} SOL</span></div>
        <div class="mb-2"><span class="font-semibold">PNL:</span> <span :class="selectedWallet.pnl >= 0 ? 'text-green-600' : 'text-red-600'">{{ selectedWallet.pnl > 0 ? '+' : '' }}{{ selectedWallet.pnl }}</span></div>
        <div class="mb-2"><span class="font-semibold">Transactions:</span> {{ selectedWallet.txCount }}</div>
        <div class="mb-2"><span class="font-semibold">Avg Tokens/Day:</span> {{ selectedWallet.avgTokensPerDay }}</div>
        <div class="mb-2"><span class="font-semibold">Avg Hold Time (days):</span> {{ selectedWallet.avgHoldTime }}</div>
        <div class="mb-2"><span class="font-semibold">Last Active:</span> {{ selectedWallet.lastActive }}</div>
        <div class="mb-2"><span class="font-semibold">Token Holdings:</span> <span class="font-mono">{{ selectedWallet.tokens.join(', ') }}</span></div>
        <div class="mb-2"><span class="font-semibold">Recent Activity:</span>
          <ul class="ml-6 text-sm list-disc">
            <li v-for="(tx, i) in selectedWallet.recentActivity" :key="i">{{ tx }}</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ProfitableSolanaWallets',
  data() {
    return {
      search: '',
      selectedWallet: null,
      filters: {
        winrate: 0,
        profitMin: '',
        profitMax: '',
        txMin: '',
        txMax: '',
        lastActiveFrom: '',
        lastActiveTo: '',
        token: ''
      },
      sortKey: 'rank',
      sortOrder: 'asc',
      wallets: [
        {
          address: '4k3Dyjzvzp8eM8b1Qw1r1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1',
          profit: 1200.5,
          pnl: 1100.2,
          txCount: 320,
          avgTokensPerDay: 3.2,
          avgHoldTime: 14,
          lastActive: '2024-06-01',
          winrate: 85,
          tokens: ['SOL', 'USDC', 'RAY'],
          recentActivity: ['Bought 100 SOL', 'Sold 50 RAY', 'Swapped 200 USDC for SOL']
        },
        {
          address: '7Ggk1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1',
          profit: 950.2,
          pnl: 900.0,
          txCount: 210,
          avgTokensPerDay: 2.1,
          avgHoldTime: 10,
          lastActive: '2024-05-30',
          winrate: 72,
          tokens: ['SOL', 'SRM'],
          recentActivity: ['Received 500 SOL', 'Sold 100 SRM']
        },
        {
          address: '9Jj1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1k1',
          profit: 800.0,
          pnl: -150.5,
          txCount: 150,
          avgTokensPerDay: 7.5,
          avgHoldTime: 2,
          lastActive: '2024-05-28',
          winrate: 60,
          tokens: ['SOL', 'USDT', 'COPE'],
          recentActivity: ['Swapped 1000 USDT for SOL', 'Bought 200 COPE']
        }
      ]
    }
  },
  computed: {
    allTokens() {
      // Unique tokens from all wallets
      const set = new Set();
      this.wallets.forEach(w => w.tokens.forEach(t => set.add(t)));
      return Array.from(set);
    },
    filteredWallets() {
      return this.wallets.filter(w => {
        if (this.search && !w.address.toLowerCase().includes(this.search.toLowerCase())) return false;
        if (this.filters.winrate && (w.winrate === undefined || w.winrate < this.filters.winrate)) return false;
        if (this.filters.profitMin !== '' && w.profit < this.filters.profitMin) return false;
        if (this.filters.profitMax !== '' && w.profit > this.filters.profitMax) return false;
        if (this.filters.txMin !== '' && w.txCount < this.filters.txMin) return false;
        if (this.filters.txMax !== '' && w.txCount > this.filters.txMax) return false;
        if (this.filters.lastActiveFrom && w.lastActive < this.filters.lastActiveFrom) return false;
        if (this.filters.lastActiveTo && w.lastActive > this.filters.lastActiveTo) return false;
        if (this.filters.token && !w.tokens.includes(this.filters.token)) return false;
        return true;
      });
    },
    sortedWallets() {
      const arr = [...this.filteredWallets];
      const key = this.sortKey;
      const order = this.sortOrder;
      arr.sort((a, b) => {
        let aValue, bValue;
        if (key === 'rank') {
          aValue = this.wallets.indexOf(a);
          bValue = this.wallets.indexOf(b);
        } else if (key === 'tokens') {
          aValue = a.tokens.join(', ');
          bValue = b.tokens.join(', ');
        } else {
          aValue = a[key];
          bValue = b[key];
        }
        if (aValue === undefined) return 1;
        if (bValue === undefined) return -1;
        if (typeof aValue === 'string' && typeof bValue === 'string') {
          return order === 'asc' ? aValue.localeCompare(bValue) : bValue.localeCompare(aValue);
        }
        return order === 'asc' ? aValue - bValue : bValue - aValue;
      });
      return arr;
    }
  },
  methods: {
    selectWallet(wallet) {
      this.selectedWallet = wallet;
    },
    sortBy(key) {
      if (this.sortKey === key) {
        this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
      } else {
        this.sortKey = key;
        this.sortOrder = 'asc';
      }
    }
  }
}
</script> 