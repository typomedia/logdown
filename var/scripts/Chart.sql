select
    Date,
    Request,
    Param,
    count (*) as Number,
    avg (Duration) as Average
from Logs
where Request = :request
  and Param = :param
  and Date like :date
group by Date, Request
order by Date, Number desc;