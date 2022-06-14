import re
import subprocess

import matplotlib
import numpy as np
from matplotlib import pyplot as plt

matplotlib.use("pgf")

matplotlib.rcParams.update({
    'pgf.texsystem': 'pdflatex',
    'font.family': 'serif',
    'text.usetex': True,
    'font.size': 11,
    'pgf.rcfonts': False
})


def rounds_until_agreement(s):
    logs = s.splitlines()
    p = re.compile(r'^.*Processor.*with\sV\s=.*$|^.*Round.*finished.*$')
    logs = [o for o in logs if p.match(o.decode('utf-8'))]
    all_v = {}
    i = 0
    for log in logs:
        # print(i, log)
        r = re.compile(r'^.*Processor.*with\sV\s=.*$')
        if r.match(log.decode('utf-8')):
            data = log.split(b' ')
            all_v[data[3]] = data[11]
        else:
            if len(all_v) > 0:
                _, consensus = next(iter(all_v.items()))
                reached = True
                for _, v in all_v.items():
                    if v != consensus:
                        reached = False
                if reached:
                    # print(consensus)
                    return i
                i += 1
                all_v.clear()


def run(protocol, n, t, v, ind, strategy):
    command = 'go run ../example/consensus_sync_{protocol}.go {n} {t} {v} {ind} {strategy}'.format(
        protocol=protocol,
        n=str(n),
        t=str(t),
        v=' '.join(map(str, v)),
        ind=' '.join(map(str, ind)),
        strategy=strategy
    )
    print(command)
    result = subprocess.run(command, check=True, capture_output=True, shell=True)
    return rounds_until_agreement(result.stderr)


#####################################################################


rb = 5
rs = 5
b_delta = 0.1


def plot1(n, t, prot, dest):
    x = []
    y_opt = []
    y_ran = []
    for i in range(1, t, 1):
        x.append(i)
        res_opt = []
        res_ran = []
        reps = 1
        if prot == 'ben_or':
            reps *= rb
        for b in np.arange(0, 0.51, b_delta):
            z = round(b * (n - i))
            v = np.concatenate((np.zeros((z,), dtype=int), np.ones((n - i - z,), dtype=int)))
            ind = np.random.choice(i, i, replace=False) + 1
            for j in sorted(ind):
                v = np.insert(v, j - 1, 0)
            for _ in range(reps):
                res_opt.append(run(prot, n, i, v, ind, 'Optimal'))
                for _ in range(rs):
                    res_ran.append(run(prot, n, i, v, ind, 'Random'))
        y_opt.append(np.mean(res_opt))
        y_ran.append(np.mean(res_ran))

    print(x)
    print(y_opt)
    print(y_ran)
    dest.scatter(x, y_opt, marker='1', label='optimal')
    dest.scatter(x, y_ran, marker='2', label='random')

    dest.set_xlabel('$t$')
    dest.set_ylabel('$p$')


def plot2(n, t, prot, dest):
    x = []
    y_opt = []
    y_ran = []
    for b in np.arange(0, 0.51, b_delta):
        x.append(b)
        res_opt = []
        res_ran = []
        reps = 1
        if prot == 'ben_or':
            reps *= rb
        for i in range(1, t, 1):
            z = round(b * (n - i))
            v = np.concatenate((np.zeros((z,), dtype=int), np.ones((n - i - z,), dtype=int)))
            ind = np.random.choice(i, i, replace=False) + 1
            for j in sorted(ind):
                v = np.insert(v, j - 1, 0)
            for _ in range(reps):
                res_opt.append(run(prot, n, i, v, ind, 'Optimal'))
                for _ in range(rs):
                    res_ran.append(run(prot, n, i, v, ind, 'Random'))
        y_opt.append(np.mean(res_opt))
        y_ran.append(np.mean(res_ran))

    print(x)
    print(y_opt)
    print(y_ran)
    dest.scatter(x, y_opt, marker='1', label='optimal')
    dest.scatter(x, y_ran, marker='2', label='random')

    dest.set_xlabel('$b$')
    dest.set_ylabel('$p$')


def plot3(n, t, strategy, dest, random_id):
    x = []
    y_sb = []
    y_pk = []
    y_bo = []
    for i in range(1, t, 1):
        x.append(i)
        res_sb = []
        res_pk = []
        res_bo = []
        reps = 1
        if strategy == 'Random':
            reps *= rs
        for b in np.arange(0, 0.51, b_delta):
            z = round(b * (n - i))
            v = np.concatenate((np.zeros((z,), dtype=int), np.ones((n - i - z,), dtype=int)))
            ind = np.random.choice(n, i, replace=False) + 1 if random_id else np.random.choice(i, i, replace=False) + 1
            for j in sorted(ind):
                v = np.insert(v, j - 1, 0)
            for _ in range(reps):
                res_sb.append(run('single_bit', n, i, v, ind, strategy))
                res_pk.append(run('phase_king', n, i, v, ind, strategy))
                for _ in range(rb):
                    res_bo.append(run('ben_or', n, i, v, ind, strategy))
        y_sb.append(np.mean(res_sb))
        y_pk.append(np.mean(res_pk))
        y_bo.append(np.mean(res_bo))

    print(x)
    print(y_sb)
    print(y_pk)
    print(y_bo)
    dest.scatter(x, y_sb, marker='1', label='Single-bit')
    dest.scatter(x, y_pk, marker='2', label='Phase king')
    dest.scatter(x, y_bo, marker='3', label='Ben-Or\'s')

    dest.set_xlabel('$t$')
    dest.set_ylabel('$p$')


def plot4(n, t, strategy, dest, random_id):
    x = []
    y_sb = []
    y_pk = []
    y_bo = []
    for b in np.arange(0, 0.51, b_delta):
        x.append(b)
        res_sb = []
        res_pk = []
        res_bo = []
        reps = 1
        if strategy == 'Random':
            reps *= rs
        for i in range(1, t, 1):
            z = round(b * (n - i))
            v = np.concatenate((np.zeros((z,), dtype=int), np.ones((n - i - z,), dtype=int)))
            ind = np.random.choice(n, i, replace=False) + 1 if random_id else np.random.choice(i, i, replace=False) + 1
            for j in sorted(ind):
                v = np.insert(v, j - 1, 0)
            for _ in range(reps):
                res_sb.append(run('single_bit', n, i, v, ind, strategy))
                res_pk.append(run('phase_king', n, i, v, ind, strategy))
                for _ in range(rb):
                    res_bo.append(run('ben_or', n, i, v, ind, strategy))
        y_sb.append(np.mean(res_sb))
        y_pk.append(np.mean(res_pk))
        y_bo.append(np.mean(res_bo))

    print(x)
    print(y_sb)
    print(y_pk)
    print(y_bo)
    dest.scatter(x, y_sb, marker='1', label='Single-bit')
    dest.scatter(x, y_pk, marker='2', label='Phase king')
    dest.scatter(x, y_bo, marker='3', label='Ben-Or\'s')

    dest.set_xlabel('$b$')
    dest.set_ylabel('$p$')


def plot12():
    n = 40
    fig, axs = plt.subplots(3, 2, figsize=(4.98, 7))

    plot1(n, n // 4, 'single_bit', axs[0][0])
    axs[0][0].set_title('Single-bit')

    plot1(n, n // 3, 'phase_king', axs[1][0])
    axs[1][0].set_title('Phase king')

    plot1(n, n // 5, 'ben_or', axs[2][0])
    axs[2][0].set_title('Ben-Or\'s')

    plot2(n, n // 4, 'single_bit', axs[0][1])
    axs[0][1].set_title('Single-bit')

    plot2(n, n // 3, 'phase_king', axs[1][1])
    axs[1][1].set_title('Phase king')

    plot2(n, n // 5, 'ben_or', axs[2][1])
    axs[2][1].set_title('Ben-Or\'s')

    handle, label = axs[0][0].get_legend_handles_labels()
    fig.legend(handle, label, loc=(0.27, 0.95), ncol=2)

    fig.tight_layout()
    fig.subplots_adjust(top=0.9)

    plt.savefig('plot12.png')
    plt.savefig('plot12.pgf')


def plot34(random_id):
    n = 40
    fig, axs = plt.subplots(1, 2, figsize=(4.98, 2.70))

    plot3(n, n // 5, 'Random', axs[0], random_id)

    plot4(n, n // 5, 'Random', axs[1], random_id)

    handle, label = axs[0].get_legend_handles_labels()
    fig.legend(handle, label, loc=(0.10, 0.8), ncol=3)

    fig.tight_layout()
    fig.subplots_adjust(top=0.75)

    plt.savefig('plot34' + str(random_id) + '.png')
    plt.savefig('plot34' + str(random_id) + '.pgf')


def plot56():
    n = 40
    fig, axs = plt.subplots(1, 2, figsize=(4.98, 2.70))

    plot3(n, n // 5, 'Optimal', axs[0], False)

    plot4(n, n // 5, 'Optimal', axs[1], False)

    handle, label = axs[0].get_legend_handles_labels()
    fig.legend(handle, label, loc=(0.10, 0.8), ncol=3)

    fig.tight_layout()
    fig.subplots_adjust(top=0.75)

    plt.savefig('plot56.png')
    plt.savefig('plot56.pgf')


plot12()
plot34(True)
plot34(False)
plot56()
