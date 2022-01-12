<?php

namespace AppBundle\Controller;

use AppBundle\Helper\DateHelper;
use AppBundle\Repository\LogRepository;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

/**
 * @Route("/")
 */
class LogController extends AbstractController
{
    /**
     * @var array $data
     */
    protected $data;

    /**
     * @var LogRepository $repo
     */
    private $repo;

    public function __construct(LogRepository $repo)
    {
        $this->repo = $repo;
    }

    /**
     * @Route("/", name="logs_index", methods={"GET"})
     */
    public function index(Request $request): Response
    {
        $logs = $this->repo->get('Request.sql');
        $search = trim($request->query->get('search'));

        return $this->render('@App/logs/index.html.twig', [
            'search' => $search,
            'logs' => $logs,
        ]);
    }

    /**
     * @Route("/search", name="logs_search", methods={"GET"})
     */
    public function search(Request $request): Response
    {
        $logs = $this->repo->get('Search.sql');
        $view = DateHelper::analyzeDate($request->query->get('date'));

        return $this->render('@App/search/index.html.twig', [
            'logs' => $logs,
            'view'   => $view
        ]);
    }
}
