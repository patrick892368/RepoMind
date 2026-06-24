<?php

namespace App\Controller;

use Symfony\Component\Routing\Attribute\Route;

#[Route('/blog')]
class BlogController
{
    #[Route('/', name: 'blog_index', methods: ['GET'])]
    #[Route('/rss.xml', name: 'blog_rss', methods: ['GET'])]
    public function index(): Response
    {
    }

    #[Route('/posts/{slug:post}', name: 'blog_post', methods: ['GET'])]
    public function postShow(): Response
    {
    }

    #[Route('/comment/{postSlug}/new', name: 'comment_new', methods: ['POST'])]
    public function commentNew(): Response
    {
    }
}
