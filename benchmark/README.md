
## Benchmarks

Read + Write time [Sec] Banchmark : 1000 writes + 1000 reads

#### Performance

When run on desktop, cpu and memory usage are also recorded. On VM, see the [Performance](/benchmark/PERF.md) doc.

#### Run on desktop:

| Backend  | Time       | %CPU      | RSS byte      |
|----------|------------|-----------|---------------|
|memory    |  0m2.011s  | 0.2 - 5.5 | 7456 - 11028  |
|mongo     |  0m4.885s  | 0.5 - 0.8 | 11892 - 11892 |
|sqlite3   |  0m14.471s | 0.2 - 7.4 | 8416 - 12560  |

#### Run on VM running OpenShift:

| DB/Backend          | Time        |
|---------------------|-------------|
|Hawkular/Casandra    |  2m8.783s   |
|Mohawk/Memory        |  0m22.833s  |

###### Benchmark real run time chart

![Time chart](/benchmark/time-vm.png?raw=true "benchmark time vm")

###### Benchmark real run time chart

![Time chart](/benchmark/time.png?raw=true "benchmark time")

###### Benchmark cpu usage chart

![CPU chart](/benchmark/cpu.png?raw=true "benchmark cpu")

###### Benchmark memory usage chart

![Mem chart](/benchmark/mem.png?raw=true "benchmark mem")

#### Backend-Mongo

```
$ date; time ./benchmark.py; date
```
```
Tue Jun 27 23:41:45 IDT 2017
{u'MohawkVersion': u'0.22.1', u'MohawkBackend': u'Backend-Mongo', u'MetricsService': u'STARTED', u'Implementation-Version': u'0.21.0'}

real	0m4.885s
user	0m1.257s
sys	0m0.483s
Tue Jun 27 23:41:50 IDT 2017
```

```
$ while sleep 0.5; do  echo $(date) $(ps -p $(pidof mohawk) -o pcpu= -o rss=) ; done;
```
```
Tue Jun 27 23:41:45 IDT 2017 0.5 11892
Tue Jun 27 23:41:45 IDT 2017 0.6 11892
Tue Jun 27 23:41:46 IDT 2017 0.6 11892
Tue Jun 27 23:41:47 IDT 2017 0.7 11892
Tue Jun 27 23:41:47 IDT 2017 0.7 11892
Tue Jun 27 23:41:48 IDT 2017 0.7 11892
Tue Jun 27 23:41:48 IDT 2017 0.7 11892
Tue Jun 27 23:41:49 IDT 2017 0.8 11892
Tue Jun 27 23:41:49 IDT 2017 0.8 11892
Tue Jun 27 23:41:50 IDT 2017 0.8 11892
Tue Jun 27 23:41:50 IDT 2017 0.8 11892
Tue Jun 27 23:41:51 IDT 2017 0.8 11892
```

#### Backend-Sqlite3

```
$ date; time ./benchmark.py; date
```
```
Tue Jun 27 23:43:36 IDT 2017
{u'MohawkVersion': u'0.22.1', u'MohawkBackend': u'Backend-Sqlite3', u'MetricsService': u'STARTED', u'Implementation-Version': u'0.21.0'}

real	0m14.471s
user	0m1.144s
sys	0m0.443s
Tue Jun 27 23:43:50 IDT 2017
```

```
$ while sleep 0.5; do  echo $(date) $(ps -p $(pidof mohawk) -o pcpu= -o rss=) ; done;
```
```
Tue Jun 27 23:43:36 IDT 2017 0.2 8416
Tue Jun 27 23:43:36 IDT 2017 0.7 8640
Tue Jun 27 23:43:37 IDT 2017 1.3 8936
Tue Jun 27 23:43:37 IDT 2017 1.7 9592
Tue Jun 27 23:43:38 IDT 2017 2.0 9800
Tue Jun 27 23:43:38 IDT 2017 2.5 10012
Tue Jun 27 23:43:39 IDT 2017 2.8 10208
Tue Jun 27 23:43:40 IDT 2017 3.2 10604
Tue Jun 27 23:43:40 IDT 2017 3.3 10868
Tue Jun 27 23:43:41 IDT 2017 3.8 11284
Tue Jun 27 23:43:41 IDT 2017 3.8 11704
Tue Jun 27 23:43:42 IDT 2017 4.2 11704
Tue Jun 27 23:43:42 IDT 2017 4.2 11892
Tue Jun 27 23:43:43 IDT 2017 4.6 11892
Tue Jun 27 23:43:43 IDT 2017 4.7 11892
Tue Jun 27 23:43:44 IDT 2017 5.0 11892
Tue Jun 27 23:43:44 IDT 2017 5.0 11892
Tue Jun 27 23:43:45 IDT 2017 5.3 12176
Tue Jun 27 23:43:45 IDT 2017 5.3 12176
Tue Jun 27 23:43:46 IDT 2017 5.6 12176
Tue Jun 27 23:43:46 IDT 2017 5.6 12176
Tue Jun 27 23:43:47 IDT 2017 5.8 12176
Tue Jun 27 23:43:47 IDT 2017 5.8 12176
Tue Jun 27 23:43:48 IDT 2017 6.0 12376
Tue Jun 27 23:43:48 IDT 2017 6.0 12560
Tue Jun 27 23:43:49 IDT 2017 6.2 12560
Tue Jun 27 23:43:49 IDT 2017 6.8 12560
Tue Jun 27 23:43:50 IDT 2017 7.4 12560
Tue Jun 27 23:43:51 IDT 2017 7.4 12560
```

#### Backend-Memory

```
$ date; time ./benchmark.py; date
```
```
Tue Jun 27 23:46:50 IDT 2017
{u'MohawkVersion': u'0.22.1', u'MohawkBackend': u'Backend-Memory', u'MetricsService': u'STARTED', u'Implementation-Version': u'0.21.0'}

real	0m2.011s
user	0m1.108s
sys	0m0.437s
Tue Jun 27 23:46:52 IDT 2017
```

```
$ while sleep 0.5; do  echo $(date) $(ps -p $(pidof mohawk) -o pcpu= -o rss=) ; done;
```
```
Tue Jun 27 23:46:50 IDT 2017 0.2 7456
Tue Jun 27 23:46:50 IDT 2017 2.2 10432
Tue Jun 27 23:46:51 IDT 2017 4.1 10824
Tue Jun 27 23:46:51 IDT 2017 5.1 11028
Tue Jun 27 23:46:52 IDT 2017 6.1 11028
Tue Jun 27 23:46:52 IDT 2017 5.5 11028
Tue Jun 27 23:46:53 IDT 2017 5.5 11028
```

#### Hawkular/Casandra vs. Mohawk/Memory

This benchmark run on two identical vms running OpenShift with identical load.
The benchmark was done using the OpenShift metric engine.

###### Hawkular/Casandra

```
[root@yzamir-centos7-1 ~]# time ./test.py
```
```
{u'Cassandra': u'up', u'MetricsService': u'STARTED', u'Implementation-Version': u'0.26.1.Final', u'Built-From-Git-SHA1': u'45b148c834ed62018f153c23187b4436ae4208fe'}

real	2m8.783s
user	0m9.555s
sys	0m0.785s
```

###### Mohawk/Memory

```
[root@yzamir-centos7-2 ~]# time ./test.py
```
```
{u'MohawkVersion': u'0.15.3', u'MohawkBackend': u'Backend-Memory', u'MetricsService': u'STARTED', u'Implementation-Version': u'0.21.0'}

real	0m22.833s
user	0m8.508s
sys	0m0.627s
```
