select
    strftime("%Y-%m", Date) as Month,
       Method,
       Request,
       Param,
       Port,
       Status,
       avg(Duration) as Duration,
       count(*)      as Number
from Logs
--where Date like '2021-11-%'
   where (Request like :search
   or Param like :search)
group by Month, Method, Request, Param, Port, Status
order by Month desc, Number desc
limit 500;